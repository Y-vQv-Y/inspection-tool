package commands

import (
	"fmt"
	"inspection-tool/internal/k8s"
	"inspection-tool/internal/server"
	"inspection-tool/internal/ssh"
	"inspection-tool/pkg/models"
	"inspection-tool/pkg/report"
	"inspection-tool/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// AllOptions 综合巡检选项
type AllOptions struct {
	// K8s相关
	Kubeconfig string
	Namespaces string

	// 服务器相关
	Hosts       string
	SSHUser     string
	SSHPassword string
	SSHPort     int

	// 报告相关
	Output   string
	Format   string
	Detailed bool
}

// NewAllCommand 创建综合巡检命令
func NewAllCommand() *cobra.Command {
	opts := &AllOptions{}

	cmd := &cobra.Command{
		Use:   "all",
		Short: "执行综合巡检(服务器+K8s)",
		Long:  `同时执行服务器和Kubernetes集群的综合巡检。`,
		Example: `  # K8s集群巡检(会自动巡检worker节点)
  inspection-tool all --kubeconfig ~/.kube/config --ssh-user root --ssh-password pass

  # 同时巡检指定服务器和K8s集群
  inspection-tool all --kubeconfig ~/.kube/config --hosts "192.168.1.10,192.168.1.11" --ssh-user root --ssh-password pass`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAllInspection(opts)
		},
	}

	homeDir, _ := os.UserHomeDir()
	defaultKubeconfig := filepath.Join(homeDir, ".kube", "config")

	cmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", defaultKubeconfig, "kubeconfig文件路径")
	cmd.Flags().StringVar(&opts.Namespaces, "namespaces", "", "要检查的命名空间(逗号分隔)")
	cmd.Flags().StringVar(&opts.Hosts, "hosts", "", "额外的服务器地址(逗号分隔)")
	cmd.Flags().StringVar(&opts.SSHUser, "ssh-user", "root", "SSH用户名")
	cmd.Flags().StringVar(&opts.SSHPassword, "ssh-password", "", "SSH密码")
	cmd.Flags().IntVar(&opts.SSHPort, "ssh-port", 22, "SSH端口")
	cmd.Flags().StringVar(&opts.Output, "output", "./reports", "报告输出目录")
	cmd.Flags().StringVar(&opts.Format, "format", "json", "报告格式(json/yaml)")
	cmd.Flags().BoolVar(&opts.Detailed, "detailed", true, "生成详细报告")

	return cmd
}

func runAllInspection(opts *AllOptions) error {
	fmt.Println("========================================")
	fmt.Println("开始综合巡检")
	fmt.Println("========================================")
	fmt.Printf("时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("========================================\n")

	fullReport := &models.InspectionReport{
		Timestamp: time.Now(),
		Type:      "all",
	}

	// 1. 执行K8s巡检
	fmt.Println("【1/2】Kubernetes集群巡检")
	fmt.Println("----------------------------------------")
	
	var namespaces []string
	if opts.Namespaces != "" {
		namespaces = strings.Split(opts.Namespaces, ",")
		for i := range namespaces {
			namespaces[i] = strings.TrimSpace(namespaces[i])
		}
	}

	k8sInspector, err := k8s.NewInspector(&k8s.InspectorConfig{
		Kubeconfig: opts.Kubeconfig,
		Namespaces: namespaces,
		Timeout:    60 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("创建K8s巡检器失败: %w", err)
	}

	k8sReport, err := k8sInspector.Inspect()
	if err != nil {
		fmt.Printf("警告: K8s巡检失败: %v\n\n", err)
	} else {
		fullReport.K8sReport = k8sReport
		fmt.Printf("✓ K8s巡检完成: 发现 %d 个问题\n\n", len(k8sReport.Issues))
	}

	// 2. 执行服务器巡检
	fmt.Println("【2/2】服务器资源巡检")
	fmt.Println("----------------------------------------")

	// 收集要巡检的主机列表
	hosts := collectHosts(opts, k8sReport)
	
	if len(hosts) == 0 {
		fmt.Println("警告: 没有可巡检的服务器\n")
	} else {
		fmt.Printf("准备巡检 %d 台服务器\n", len(hosts))
		
		// 并发巡检服务器
		serverReports := inspectServersParallel(hosts, opts)
		
		if len(serverReports) > 0 {
			// 合并第一个服务器报告到完整报告
			// 实际应用中可能需要更复杂的合并逻辑
			fullReport.ServerReport = serverReports[0]
			
			totalIssues := 0
			for _, sr := range serverReports {
				totalIssues += len(sr.Issues)
			}
			fmt.Printf("✓ 服务器巡检完成: 发现 %d 个问题\n\n", totalIssues)
		}
	}

	// 构建摘要
	utils.BuildInspectionSummary(fullReport)

	// 生成报告
	fmt.Println("正在生成综合报告...")
	generator := report.NewGenerator(opts.Format, opts.Output, opts.Detailed)
	reportPath, err := generator.GenerateFullReport(fullReport)
	if err != nil {
		return fmt.Errorf("生成报告失败: %w", err)
	}

	// 打印摘要
	report.PrintSummary(fullReport)

	fmt.Printf("\n✓ 综合报告已保存: %s\n", reportPath)

	// 根据严重问题返回退出码
	if fullReport.Summary.CriticalIssues > 0 {
		return fmt.Errorf("发现 %d 个严重问题", fullReport.Summary.CriticalIssues)
	}

	return nil
}

// collectHosts 收集要巡检的主机列表
func collectHosts(opts *AllOptions, k8sReport *models.K8sReport) []string {
	hostsMap := make(map[string]bool)

	// 从命令行参数添加
	if opts.Hosts != "" {
		hosts := strings.Split(opts.Hosts, ",")
		for _, host := range hosts {
			host = strings.TrimSpace(host)
			if host != "" {
				hostsMap[host] = true
			}
		}
	}

	// 从K8s节点添加(如果提供了SSH密码)
	if opts.SSHPassword != "" && k8sReport != nil {
		for _, node := range k8sReport.Nodes {
			// 尝试从节点获取IP地址
			// 实际应该从node.Status.Addresses中获取InternalIP
			// 这里简化处理
			if node.Name != "" {
				hostsMap[node.Name] = true
			}
		}
	}

	// 转换为切片
	hosts := make([]string, 0, len(hostsMap))
	for host := range hostsMap {
		hosts = append(hosts, host)
	}

	return hosts
}

// inspectServersParallel 并发巡检服务器
func inspectServersParallel(hosts []string, opts *AllOptions) []*models.ServerReport {
	var wg sync.WaitGroup
	var mu sync.Mutex
	reports := make([]*models.ServerReport, 0)

	// 限制并发数
	semaphore := make(chan struct{}, 5)

	for _, host := range hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			fmt.Printf("  → 正在巡检: %s\n", h)

			// 创建SSH客户端
			sshClient, err := ssh.NewClient(&ssh.Config{
				Host:     h,
				Port:     opts.SSHPort,
				User:     opts.SSHUser,
				Password: opts.SSHPassword,
				Timeout:  30 * time.Second,
			})
			if err != nil {
				fmt.Printf("  ✗ %s: SSH连接失败 - %v\n", h, err)
				return
			}
			defer sshClient.Close()

			// 创建巡检器
			inspector, err := server.NewInspector(sshClient)
			if err != nil {
				fmt.Printf("  ✗ %s: 创建巡检器失败 - %v\n", h, err)
				return
			}
			defer inspector.Close()

			// 执行巡检
			serverReport, err := inspector.Inspect()
			if err != nil {
				fmt.Printf("  ✗ %s: 巡检失败 - %v\n", h, err)
				return
			}

			mu.Lock()
			reports = append(reports, serverReport)
			mu.Unlock()

			fmt.Printf("  ✓ %s: 完成 (%d个问题)\n", h, len(serverReport.Issues))
		}(host)
	}

	wg.Wait()
	return reports
}
