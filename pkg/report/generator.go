package report

import (
	"encoding/json"
	"fmt"
	"inspection-tool/pkg/models"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Generator 报告生成器
type Generator struct {
	format    string // json, yaml
	outputDir string
	detailed  bool
}

// NewGenerator 创建报告生成器
func NewGenerator(format, outputDir string, detailed bool) *Generator {
	return &Generator{
		format:    format,
		outputDir: outputDir,
		detailed:  detailed,
	}
}

// GenerateServerReport 生成服务器巡检报告
func (g *Generator) GenerateServerReport(report *models.ServerReport) (string, error) {
	filename := fmt.Sprintf("server_%s_%s.%s",
		report.Host,
		report.Timestamp.Format("20060102_150405"),
		g.format,
	)

	return g.saveReport(filename, report)
}

// GenerateK8sReport 生成K8s巡检报告
func (g *Generator) GenerateK8sReport(report *models.K8sReport) (string, error) {
	filename := fmt.Sprintf("k8s_%s.%s",
		report.Timestamp.Format("20060102_150405"),
		g.format,
	)

	return g.saveReport(filename, report)
}

// GenerateFullReport 生成完整巡检报告
func (g *Generator) GenerateFullReport(report *models.InspectionReport) (string, error) {
	filename := fmt.Sprintf("inspection_%s_%s.%s",
		report.Type,
		report.Timestamp.Format("20060102_150405"),
		g.format,
	)

	return g.saveReport(filename, report)
}

// saveReport 保存报告
func (g *Generator) saveReport(filename string, data interface{}) (string, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	filepath := filepath.Join(g.outputDir, filename)

	var content []byte
	var err error

	switch g.format {
	case "json":
		if g.detailed {
			content, err = json.MarshalIndent(data, "", "  ")
		} else {
			content, err = json.Marshal(data)
		}
	case "yaml":
		content, err = yaml.Marshal(data)
	default:
		return "", fmt.Errorf("unsupported format: %s", g.format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filepath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	return filepath, nil
}

// PrintSummary 打印摘要
func PrintSummary(report *models.InspectionReport) {
	fmt.Println("\n========================================")
	fmt.Println("巡检报告摘要")
	fmt.Println("========================================")
	fmt.Printf("巡检类型: %s\n", report.Type)
	fmt.Printf("巡检时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Println("========================================")
	
	fmt.Printf("总问题数: %d\n", report.Summary.TotalIssues)
	fmt.Printf("  - 严重: %d\n", report.Summary.CriticalIssues)
	fmt.Printf("  - 警告: %d\n", report.Summary.WarningIssues)
	fmt.Printf("  - 信息: %d\n", report.Summary.InfoIssues)
	fmt.Printf("整体状态: %s\n", report.Summary.Status)
	fmt.Println("========================================")

	if len(report.Summary.Messages) > 0 {
		fmt.Println("\n关键问题:")
		for i, msg := range report.Summary.Messages {
			if i >= 10 {
				fmt.Printf("... 还有 %d 条问题\n", len(report.Summary.Messages)-10)
				break
			}
			fmt.Printf("  %d. %s\n", i+1, msg)
		}
		fmt.Println("========================================")
	}
}

// PrintServerSummary 打印服务器巡检摘要
func PrintServerSummary(report *models.ServerReport) {
	fmt.Println("\n========================================")
	fmt.Println("服务器巡检摘要")
	fmt.Println("========================================")
	fmt.Printf("主机: %s\n", report.Host)
	fmt.Printf("操作系统: %s %s\n", report.OS.Family, report.OS.Version)
	fmt.Printf("内核版本: %s\n", report.OS.KernelVer)
	fmt.Printf("运行时间: %d秒 (%.1f天)\n", report.OS.Uptime, float64(report.OS.Uptime)/86400)
	fmt.Println("========================================")

	fmt.Println("\nCPU:")
	fmt.Printf("  核心数: %d\n", report.CPU.CoreCount)
	fmt.Printf("  负载: %.2f / %.2f / %.2f\n", report.CPU.Load1, report.CPU.Load5, report.CPU.Load15)
	fmt.Printf("  使用率: %.2f%%\n", report.CPU.UsagePercent)
	fmt.Printf("  IO等待: %.2f%%\n", report.CPU.IowaitPercent)

	fmt.Println("\n内存:")
	fmt.Printf("  总量: %d MB\n", report.Memory.TotalMB)
	fmt.Printf("  已用: %d MB (%.2f%%)\n", report.Memory.UsedMB, report.Memory.UsagePercent)
	fmt.Printf("  可用: %d MB\n", report.Memory.AvailableMB)
	fmt.Printf("  Swap: %d / %d MB (%.2f%%)\n", report.Memory.SwapUsedMB, report.Memory.SwapTotalMB, report.Memory.SwapPercent)

	fmt.Println("\n磁盘:")
	for _, disk := range report.Disk {
		fmt.Printf("  %s (%s): %.2f%% 使用\n", disk.MountPoint, disk.Device, disk.UsagePercent)
	}

	fmt.Println("\n网络:")
	fmt.Printf("  TCP连接: ESTABLISHED=%d, TIME_WAIT=%d\n",
		report.Network.TCPConnections.Established,
		report.Network.TCPConnections.TimeWait)

	fmt.Println("\n系统:")
	fmt.Printf("  文件句柄: %d / %d (%.2f%%)\n",
		report.System.FileHandlesAllocated,
		report.System.FileHandlesMax,
		report.System.FileHandlesPercent)
	fmt.Printf("  进程数: %d\n", report.System.ProcessCount)

	if len(report.Issues) > 0 {
		fmt.Println("\n问题列表:")
		for i, issue := range report.Issues {
			if i >= 10 {
				fmt.Printf("... 还有 %d 个问题\n", len(report.Issues)-10)
				break
			}
			fmt.Printf("  [%s] %s: %s\n", issue.Level, issue.Category, issue.Message)
		}
	}
	fmt.Println("========================================")
}

// PrintK8sSummary 打印K8s巡检摘要
func PrintK8sSummary(report *models.K8sReport) {
	fmt.Println("\n========================================")
	fmt.Println("Kubernetes巡检摘要")
	fmt.Println("========================================")
	fmt.Printf("集群版本: %s\n", report.ClusterInfo.Version)
	fmt.Printf("节点数: %d\n", report.ClusterInfo.NodeCount)
	fmt.Printf("命名空间数: %d\n", report.ClusterInfo.NamespaceCount)
	fmt.Printf("Pod总数: %d\n", report.ClusterInfo.PodCount)
	fmt.Println("========================================")

	fmt.Println("\n节点状态:")
	readyCount := 0
	for _, node := range report.Nodes {
		if node.Ready {
			readyCount++
			fmt.Printf("  ✓ %s: Ready (CPU: %.1f%%, Memory: %.1f%%, Pods: %d/%d)\n",
				node.Name, node.CPUPercent, node.MemoryPercent, node.PodCount, node.PodsCapacity)
		} else {
			fmt.Printf("  ✗ %s: NotReady\n", node.Name)
		}
	}
	fmt.Printf("  Ready: %d / %d\n", readyCount, len(report.Nodes))

	fmt.Println("\n控制平面:")
	fmt.Printf("  API Server: %s\n", getHealthStatus(report.APIServerStatus.Healthy))
	fmt.Printf("  etcd: %s (成员: %d)\n", getHealthStatus(report.EtcdStatus.Healthy), report.EtcdStatus.ClusterSize)
	fmt.Printf("  Controller Manager: %s\n", getHealthStatus(report.ControllerStatus.Healthy))
	fmt.Printf("  Scheduler: %s\n", getHealthStatus(report.SchedulerStatus.Healthy))

	if len(report.Issues) > 0 {
		fmt.Println("\n问题列表:")
		for i, issue := range report.Issues {
			if i >= 10 {
				fmt.Printf("... 还有 %d 个问题\n", len(report.Issues)-10)
				break
			}
			fmt.Printf("  [%s] %s: %s\n", issue.Level, issue.Category, issue.Message)
		}
	}
	fmt.Println("========================================")
}

// getHealthStatus 获取健康状态字符串
func getHealthStatus(healthy bool) string {
	if healthy {
		return "✓ Healthy"
	}
	return "✗ Unhealthy"
}

// CleanupOldReports 清理旧报告
func CleanupOldReports(outputDir string, retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			filepath := filepath.Join(outputDir, entry.Name())
			os.Remove(filepath)
		}
	}

	return nil
}
