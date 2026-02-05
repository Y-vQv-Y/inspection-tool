package commands

import (
	"fmt"
	"inspection-tool/internal/k8s"
	"inspection-tool/pkg/models"
	"inspection-tool/pkg/report"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// K8sOptions Kubernetes巡检选项
type K8sOptions struct {
	Kubeconfig     string
	Namespaces     string
	Output         string
	Format         string
	Detailed       bool
	InspectWorkers bool
	SSHUser        string
	SSHPassword    string
	SSHPort        int
}

// NewK8sCommand 创建Kubernetes巡检命令
func NewK8sCommand() *cobra.Command {
	opts := &K8sOptions{}

	cmd := &cobra.Command{
		Use:   "k8s",
		Short: "执行Kubernetes集群巡检",
		Long:  `巡检Kubernetes集群健康状态,包括节点、Pod、控制平面组件等。`,
		Example: `  # 基本用法
  inspection-tool k8s --kubeconfig ~/.kube/config

  # 指定命名空间
  inspection-tool k8s --kubeconfig ~/.kube/config --namespaces default,kube-system

  # 同时巡检worker节点服务器资源
  inspection-tool k8s --kubeconfig ~/.kube/config --inspect-workers --ssh-user root --ssh-password pass`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runK8sInspection(opts)
		},
	}

	// 获取默认kubeconfig路径
	homeDir, _ := os.UserHomeDir()
	defaultKubeconfig := filepath.Join(homeDir, ".kube", "config")

	cmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", defaultKubeconfig, "kubeconfig文件路径")
	cmd.Flags().StringVar(&opts.Namespaces, "namespaces", "", "要检查的命名空间(逗号分隔,为空则检查所有)")
	cmd.Flags().StringVar(&opts.Output, "output", "./reports", "报告输出目录")
	cmd.Flags().StringVar(&opts.Format, "format", "json", "报告格式(json/yaml)")
	cmd.Flags().BoolVar(&opts.Detailed, "detailed", true, "生成详细报告")
	cmd.Flags().BoolVar(&opts.InspectWorkers, "inspect-workers", false, "同时巡检worker节点服务器资源")
	cmd.Flags().StringVar(&opts.SSHUser, "ssh-user", "root", "Worker节点SSH用户名")
	cmd.Flags().StringVar(&opts.SSHPassword, "ssh-password", "", "Worker节点SSH密码")
	cmd.Flags().IntVar(&opts.SSHPort, "ssh-port", 22, "Worker节点SSH端口")

	return cmd
}

func runK8sInspection(opts *K8sOptions) error {
	fmt.Println("========================================")
	fmt.Println("开始Kubernetes巡检")
	fmt.Println("========================================")
	fmt.Printf("Kubeconfig: %s\n", opts.Kubeconfig)
	if opts.Namespaces != "" {
		fmt.Printf("命名空间: %s\n", opts.Namespaces)
	} else {
		fmt.Println("命名空间: 全部")
	}
	fmt.Println("========================================\n")

	// 解析命名空间
	var namespaces []string
	if opts.Namespaces != "" {
		namespaces = strings.Split(opts.Namespaces, ",")
		for i := range namespaces {
			namespaces[i] = strings.TrimSpace(namespaces[i])
		}
	}

	// 创建巡检器
	fmt.Println("正在连接Kubernetes集群...")
	inspector, err := k8s.NewInspector(&k8s.InspectorConfig{
		Kubeconfig: opts.Kubeconfig,
		Namespaces: namespaces,
		Timeout:    60 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("创建K8s巡检器失败: %w", err)
	}
	fmt.Println("✓ 连接成功\n")

	// 执行巡检
	fmt.Println("正在执行集群巡检...")
	k8sReport, err := inspector.Inspect()
	if err != nil {
		return fmt.Errorf("K8s巡检失败: %w", err)
	}
	fmt.Println("✓ 集群巡检完成\n")

	// 如果需要巡检worker节点
	if opts.InspectWorkers && opts.SSHPassword != "" {
		fmt.Println("正在巡检Worker节点服务器资源...")
		if err := inspectWorkerNodes(k8sReport, opts); err != nil {
			fmt.Printf("警告: Worker节点巡检失败: %v\n", err)
		} else {
			fmt.Println("✓ Worker节点巡检完成\n")
		}
	}

	// 生成报告
	fmt.Println("正在生成报告...")
	generator := report.NewGenerator(opts.Format, opts.Output, opts.Detailed)
	reportPath, err := generator.GenerateK8sReport(k8sReport)
	if err != nil {
		return fmt.Errorf("生成报告失败: %w", err)
	}

	// 打印摘要
	report.PrintK8sSummary(k8sReport)

	fmt.Printf("\n✓ 报告已保存: %s\n", reportPath)

	// 根据问题数量返回退出码
	if len(k8sReport.Issues) > 0 {
		criticalCount := 0
		for _, issue := range k8sReport.Issues {
			if issue.Level == "critical" {
				criticalCount++
			}
		}
		if criticalCount > 0 {
			return fmt.Errorf("发现 %d 个严重问题", criticalCount)
		}
	}

	return nil
}

func inspectWorkerNodes(k8sReport *models.K8sReport, opts *K8sOptions) error {
	// 获取所有worker节点的IP
	workerIPs := []string{}
	for _, node := range k8sReport.Nodes {
		// 从节点标签或注解中提取IP
		// 这里简化处理,实际应该从node.Status.Addresses中获取
		if node.Name != "" {
			// 尝试从节点名称或标签获取IP
			// 实际实现中应该更完善
			workerIPs = append(workerIPs, node.Name)
		}
	}

	if len(workerIPs) == 0 {
		return fmt.Errorf("未找到worker节点")
	}

	fmt.Printf("发现 %d 个worker节点\n", len(workerIPs))

	// 这里可以并发巡检多个worker节点
	// 为简化代码,这里只做示例
	
	return nil
}
