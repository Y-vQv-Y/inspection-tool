package k8s

import (
	"context"
	"fmt"
	"inspection-tool/pkg/models"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Inspector Kubernetes巡检器
type Inspector struct {
	clientset        *kubernetes.Clientset
	metricsClientset *metricsv.Clientset
	config           *InspectorConfig
}

// InspectorConfig 巡检配置
type InspectorConfig struct {
	Kubeconfig string
	Namespaces []string
	Timeout    time.Duration
}

// NewInspector 创建Kubernetes巡检器
func NewInspector(config *InspectorConfig) (*Inspector, error) {
	var restConfig *rest.Config
	var err error

	// 尝试使用kubeconfig
	if config.Kubeconfig != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	} else {
		// 尝试使用in-cluster配置
		restConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// 创建metrics clientset
	metricsClientset, err := metricsv.NewForConfig(restConfig)
	if err != nil {
		// metrics server可能未安装,不作为致命错误
		metricsClientset = nil
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Inspector{
		clientset:        clientset,
		metricsClientset: metricsClientset,
		config:           config,
	}, nil
}

// Inspect 执行巡检
func (i *Inspector) Inspect() (*models.K8sReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), i.config.Timeout)
	defer cancel()

	report := &models.K8sReport{
		Timestamp: time.Now(),
		Issues:    []models.Issue{},
	}

	// 收集集群信息
	if err := i.collectClusterInfo(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to collect cluster info: %w", err)
	}

	// 收集节点信息
	if err := i.collectNodeMetrics(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to collect node metrics: %w", err)
	}

	// 收集API Server状态
	if err := i.collectAPIServerMetrics(ctx, report); err != nil {
		// API Server检查失败不中断巡检
		report.Issues = append(report.Issues, models.Issue{
			Level:     "warning",
			Category:  "apiserver",
			Message:   "Failed to collect API Server metrics",
			Details:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	// 收集etcd状态
	if err := i.collectEtcdMetrics(ctx, report); err != nil {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "warning",
			Category:  "etcd",
			Message:   "Failed to collect etcd metrics",
			Details:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	// 收集Controller Manager状态
	if err := i.collectControllerMetrics(ctx, report); err != nil {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "info",
			Category:  "controller",
			Message:   "Failed to collect Controller Manager metrics",
			Details:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	// 收集Scheduler状态
	if err := i.collectSchedulerMetrics(ctx, report); err != nil {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "info",
			Category:  "scheduler",
			Message:   "Failed to collect Scheduler metrics",
			Details:   err.Error(),
			Timestamp: time.Now(),
		})
	}

	// 收集Pod信息
	if err := i.collectPodMetrics(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to collect pod metrics: %w", err)
	}

	// 分析问题
	i.analyzeIssues(report)

	return report, nil
}

// collectClusterInfo 收集集群信息
func (i *Inspector) collectClusterInfo(ctx context.Context, report *models.K8sReport) error {
	// 获取版本信息
	version, err := i.clientset.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	// 获取节点数量
	nodes, err := i.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 获取命名空间数量
	namespaces, err := i.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 获取Pod总数
	pods, err := i.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	report.ClusterInfo = models.ClusterInfo{
		Version:        version.GitVersion,
		NodeCount:      len(nodes.Items),
		NamespaceCount: len(namespaces.Items),
		PodCount:       len(pods.Items),
	}

	return nil
}

// collectNodeMetrics 收集节点指标
func (i *Inspector) collectNodeMetrics(ctx context.Context, report *models.K8sReport) error {
	nodes, err := i.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// 获取节点指标(如果metrics server可用)
	var nodeMetrics map[string]*metricsv1beta1.NodeMetrics
	if i.metricsClientset != nil {
		metrics, err := i.metricsClientset.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
		if err == nil {
			nodeMetrics = make(map[string]*metricsv1beta1.NodeMetrics)
			for idx := range metrics.Items {
				item := &metrics.Items[idx]
				nodeMetrics[item.Name] = item
			}
		}
	}

	// 获取所有Pods用于统计节点上的Pod数量
	pods, err := i.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	nodePodCount := make(map[string]int)
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" && pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			nodePodCount[pod.Spec.NodeName]++
		}
	}

	for _, node := range nodes.Items {
		nodeMetric := i.parseNodeMetrics(node, nodeMetrics[node.Name], nodePodCount[node.Name])
		report.Nodes = append(report.Nodes, nodeMetric)
	}

	return nil
}

// parseNodeMetrics 解析节点指标
func (i *Inspector) parseNodeMetrics(node corev1.Node, metrics *metricsv1beta1.NodeMetrics, podCount int) models.NodeMetrics {
	nm := models.NodeMetrics{
		Name:             node.Name,
		Ready:            false,
		Conditions:       []models.NodeCondition{},
		CPUCapacity:      node.Status.Capacity.Cpu().String(),
		MemoryCapacity:   node.Status.Capacity.Memory().String(),
		PodsCapacity:     int(node.Status.Capacity.Pods().Value()),
		PodCount:         podCount,
		Labels:           node.Labels,
		Taints:           []string{},
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		OSImage:          node.Status.NodeInfo.OSImage,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
	}

	// 解析状态条件
	for _, condition := range node.Status.Conditions {
		nm.Conditions = append(nm.Conditions, models.NodeCondition{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})

		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			nm.Ready = true
		}
	}

	// 解析污点
	for _, taint := range node.Spec.Taints {
		nm.Taints = append(nm.Taints, fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect))
	}

	// 解析资源使用(如果metrics可用)
	if metrics != nil {
		nm.CPUUsage = metrics.Usage.Cpu().String()
		nm.MemoryUsage = metrics.Usage.Memory().String()

		// 计算使用百分比
		cpuCapacity := node.Status.Capacity.Cpu().MilliValue()
		cpuUsage := metrics.Usage.Cpu().MilliValue()
		if cpuCapacity > 0 {
			nm.CPUPercent = float64(cpuUsage) / float64(cpuCapacity) * 100
		}

		memCapacity := node.Status.Capacity.Memory().Value()
		memUsage := metrics.Usage.Memory().Value()
		if memCapacity > 0 {
			nm.MemoryPercent = float64(memUsage) / float64(memCapacity) * 100
		}
	}

	// Pod使用百分比
	if nm.PodsCapacity > 0 {
		nm.PodPercent = float64(nm.PodCount) / float64(nm.PodsCapacity) * 100
	}

	return nm
}

// collectAPIServerMetrics 收集API Server指标
func (i *Inspector) collectAPIServerMetrics(ctx context.Context, report *models.K8sReport) error {
	// 尝试访问health端点
	result := i.clientset.Discovery().RESTClient().Get().
		AbsPath("/healthz").
		Do(ctx)

	report.APIServerStatus = models.APIServerMetrics{
		Healthy: result.Error() == nil,
		Version: report.ClusterInfo.Version,
	}

	// 这里可以添加更多API Server指标收集
	// 例如通过Prometheus metrics获取请求率、错误率、延迟等

	return nil
}

// collectEtcdMetrics 收集etcd指标
func (i *Inspector) collectEtcdMetrics(ctx context.Context, report *models.K8sReport) error {
	// 尝试获取etcd Pods
	etcdPods, err := i.clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=etcd",
	})

	if err != nil {
		return err
	}

	members := []models.EtcdMember{}
	healthyCount := 0

	for _, pod := range etcdPods.Items {
		member := models.EtcdMember{
			Name:   pod.Name,
			Status: string(pod.Status.Phase),
		}

		if pod.Status.Phase == corev1.PodRunning {
			healthyCount++
		}

		members = append(members, member)
	}

	report.EtcdStatus = models.EtcdMetrics{
		Healthy:     healthyCount > 0 && healthyCount == len(members),
		ClusterSize: len(members),
		Members:     members,
	}

	// 这里可以添加更多etcd指标收集
	// 例如通过etcd API获取leader、db大小等

	return nil
}

// collectControllerMetrics 收集Controller Manager指标
func (i *Inspector) collectControllerMetrics(ctx context.Context, report *models.K8sReport) error {
	pods, err := i.clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-controller-manager",
	})

	if err != nil {
		return err
	}

	healthy := false
	leader := ""

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			healthy = true
			leader = pod.Name
			break
		}
	}

	report.ControllerStatus = models.ControllerMetrics{
		Healthy: healthy,
		Leader:  leader,
	}

	return nil
}

// collectSchedulerMetrics 收集Scheduler指标
func (i *Inspector) collectSchedulerMetrics(ctx context.Context, report *models.K8sReport) error {
	pods, err := i.clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=kube-scheduler",
	})

	if err != nil {
		return err
	}

	healthy := false
	leader := ""

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			healthy = true
			leader = pod.Name
			break
		}
	}

	report.SchedulerStatus = models.SchedulerMetrics{
		Healthy: healthy,
		Leader:  leader,
	}

	return nil
}

// collectPodMetrics 收集Pod指标
func (i *Inspector) collectPodMetrics(ctx context.Context, report *models.K8sReport) error {
	namespaces := i.config.Namespaces
	if len(namespaces) == 0 {
		// 获取所有命名空间
		nsList, err := i.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// 获取Pod指标
	var podMetricsList map[string]map[string]*metricsv1beta1.PodMetrics
	if i.metricsClientset != nil {
		podMetricsList = make(map[string]map[string]*metricsv1beta1.PodMetrics)
		for _, ns := range namespaces {
			metrics, err := i.metricsClientset.MetricsV1beta1().PodMetricses(ns).List(ctx, metav1.ListOptions{})
			if err == nil {
				podMetricsList[ns] = make(map[string]*metricsv1beta1.PodMetrics)
				for idx := range metrics.Items {
					item := &metrics.Items[idx]
					podMetricsList[ns][item.Name] = item
				}
			}
		}
	}

	for _, ns := range namespaces {
		pods, err := i.clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		for i := range pods.Items {
			pod := &pods.Items[i]
			var metrics *metricsv1beta1.PodMetrics
			if podMetricsList != nil && podMetricsList[ns] != nil {
				metrics = podMetricsList[ns][pod.Name]
			}

			podMetric := i.parsePodMetrics(*pod, metrics)
			report.Pods = append(report.Pods, podMetric)
		}
	}

	return nil
}

// parsePodMetrics 解析Pod指标
func (i *Inspector) parsePodMetrics(pod corev1.Pod, metrics *metricsv1beta1.PodMetrics) models.PodMetrics {
	pm := models.PodMetrics{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Phase:      string(pod.Status.Phase),
		Ready:      false,
		Node:       pod.Spec.NodeName,
		Labels:     pod.Labels,
		Conditions: []models.PodCondition{},
	}

	// 计算重启次数
	restartCount := 0
	for _, cs := range pod.Status.ContainerStatuses {
		restartCount += int(cs.RestartCount)
		if cs.Ready {
			pm.Ready = true
		}
	}
	pm.RestartCount = restartCount

	// 解析资源请求和限制
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				pm.CPURequest = cpu.String()
			}
			if mem := container.Resources.Requests.Memory(); mem != nil {
				pm.MemoryRequest = mem.String()
			}
		}
		if container.Resources.Limits != nil {
			if cpu := container.Resources.Limits.Cpu(); cpu != nil {
				pm.CPULimit = cpu.String()
			}
			if mem := container.Resources.Limits.Memory(); mem != nil {
				pm.MemoryLimit = mem.String()
			}
		}
	}

	// 解析资源使用
	if metrics != nil {
		for _, container := range metrics.Containers {
			if cpu := container.Usage.Cpu(); cpu != nil {
				pm.CPUUsage = cpu.String()
			}
			if mem := container.Usage.Memory(); mem != nil {
				pm.MemoryUsage = mem.String()
			}
		}
	}

	// 解析条件
	for _, condition := range pod.Status.Conditions {
		pm.Conditions = append(pm.Conditions, models.PodCondition{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	// 计算年龄
	if !pod.CreationTimestamp.IsZero() {
		pm.Age = int64(time.Since(pod.CreationTimestamp.Time).Seconds())
	}

	return pm
}

// analyzeIssues 分析问题
func (i *Inspector) analyzeIssues(report *models.K8sReport) {
	// 节点问题分析
	notReadyNodes := 0
	for _, node := range report.Nodes {
		if !node.Ready {
			notReadyNodes++
			report.Issues = append(report.Issues, models.Issue{
				Level:     "critical",
				Category:  "node",
				Message:   fmt.Sprintf("Node not ready: %s", node.Name),
				Details:   getNodeConditionDetails(node.Conditions),
				Timestamp: time.Now(),
				Suggestion: "Check node status and kubelet logs",
			})
		}

		// CPU使用率检查
		if node.CPUPercent > 80 {
			report.Issues = append(report.Issues, models.Issue{
				Level:     "warning",
				Category:  "node",
				Message:   fmt.Sprintf("High CPU usage on node %s: %.2f%%", node.Name, node.CPUPercent),
				Details:   fmt.Sprintf("CPU: %s / %s", node.CPUUsage, node.CPUCapacity),
				Timestamp: time.Now(),
				Suggestion: "Consider scaling or optimizing workloads",
			})
		}

		// 内存使用率检查
		if node.MemoryPercent > 85 {
			report.Issues = append(report.Issues, models.Issue{
				Level:     "warning",
				Category:  "node",
				Message:   fmt.Sprintf("High memory usage on node %s: %.2f%%", node.Name, node.MemoryPercent),
				Details:   fmt.Sprintf("Memory: %s / %s", node.MemoryUsage, node.MemoryCapacity),
				Timestamp: time.Now(),
				Suggestion: "Check memory-intensive pods or add more nodes",
			})
		}

		// Pod容量检查
		if node.PodPercent > 90 {
			report.Issues = append(report.Issues, models.Issue{
				Level:     "warning",
				Category:  "node",
				Message:   fmt.Sprintf("Pod capacity near limit on node %s: %.2f%%", node.Name, node.PodPercent),
				Details:   fmt.Sprintf("Pods: %d / %d", node.PodCount, node.PodsCapacity),
				Timestamp: time.Now(),
				Suggestion: "Increase max pods or add more nodes",
			})
		}
	}

	// API Server检查
	if !report.APIServerStatus.Healthy {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "critical",
			Category:  "apiserver",
			Message:   "API Server unhealthy",
			Timestamp: time.Now(),
			Suggestion: "Check API Server logs and status",
		})
	}

	// etcd检查
	if !report.EtcdStatus.Healthy {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "critical",
			Category:  "etcd",
			Message:   "etcd cluster unhealthy",
			Details:   fmt.Sprintf("Healthy members: %d / %d", countHealthyEtcdMembers(report.EtcdStatus.Members), report.EtcdStatus.ClusterSize),
			Timestamp: time.Now(),
			Suggestion: "Check etcd cluster status and logs",
		})
	}

	// Controller Manager检查
	if !report.ControllerStatus.Healthy {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "critical",
			Category:  "controller",
			Message:   "Controller Manager unhealthy",
			Timestamp: time.Now(),
			Suggestion: "Check Controller Manager logs",
		})
	}

	// Scheduler检查
	if !report.SchedulerStatus.Healthy {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "critical",
			Category:  "scheduler",
			Message:   "Scheduler unhealthy",
			Timestamp: time.Now(),
			Suggestion: "Check Scheduler logs",
		})
	}

	// Pod问题分析
	crashLoopPods := 0
	pendingPods := 0
	highRestartPods := 0

	for _, pod := range report.Pods {
		// CrashLoopBackOff检查
		if strings.Contains(pod.Phase, "CrashLoopBackOff") {
			crashLoopPods++
		}

		// Pending状态检查
		if pod.Phase == "Pending" && pod.Age > 300 { // 5分钟以上
			pendingPods++
			report.Issues = append(report.Issues, models.Issue{
				Level:     "warning",
				Category:  "pod",
				Message:   fmt.Sprintf("Pod stuck in Pending: %s/%s", pod.Namespace, pod.Name),
				Details:   getPodConditionDetails(pod.Conditions),
				Timestamp: time.Now(),
				Suggestion: "Check resource availability and scheduling constraints",
			})
		}

		// 高重启次数检查
		if pod.RestartCount > 5 {
			highRestartPods++
			report.Issues = append(report.Issues, models.Issue{
				Level:     "warning",
				Category:  "pod",
				Message:   fmt.Sprintf("High restart count: %s/%s (%d restarts)", pod.Namespace, pod.Name, pod.RestartCount),
				Timestamp: time.Now(),
				Suggestion: "Check pod logs for errors",
			})
		}

		// Not Ready检查
		if !pod.Ready && pod.Phase == "Running" {
			report.Issues = append(report.Issues, models.Issue{
				Level:     "warning",
				Category:  "pod",
				Message:   fmt.Sprintf("Pod not ready: %s/%s", pod.Namespace, pod.Name),
				Details:   getPodConditionDetails(pod.Conditions),
				Timestamp: time.Now(),
				Suggestion: "Check readiness probe and application status",
			})
		}
	}

	if crashLoopPods > 0 {
		report.Issues = append(report.Issues, models.Issue{
			Level:     "critical",
			Category:  "pod",
			Message:   fmt.Sprintf("%d pods in CrashLoopBackOff state", crashLoopPods),
			Timestamp: time.Now(),
			Suggestion: "Investigate failing pods",
		})
	}
}

// getNodeConditionDetails 获取节点条件详情
func getNodeConditionDetails(conditions []models.NodeCondition) string {
	var details []string
	for _, condition := range conditions {
		if condition.Status != "True" && condition.Type == "Ready" {
			details = append(details, fmt.Sprintf("%s: %s (%s)", condition.Type, condition.Status, condition.Reason))
		}
	}
	return strings.Join(details, "; ")
}

// getPodConditionDetails 获取Pod条件详情
func getPodConditionDetails(conditions []models.PodCondition) string {
	var details []string
	for _, condition := range conditions {
		if condition.Status == "False" {
			details = append(details, fmt.Sprintf("%s: %s (%s)", condition.Type, condition.Status, condition.Reason))
		}
	}
	return strings.Join(details, "; ")
}

// countHealthyEtcdMembers 统计健康的etcd成员
func countHealthyEtcdMembers(members []models.EtcdMember) int {
	count := 0
	for _, member := range members {
		if member.Status == "Running" {
			count++
		}
	}
	return count
}
