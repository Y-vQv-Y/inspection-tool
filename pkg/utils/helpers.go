package utils

import (
	"fmt"
	"inspection-tool/pkg/models"
	"time"
)

// BuildInspectionSummary 构建巡检摘要
func BuildInspectionSummary(report *models.InspectionReport) {
	summary := &report.Summary
	summary.TotalIssues = 0
	summary.CriticalIssues = 0
	summary.WarningIssues = 0
	summary.InfoIssues = 0
	summary.Messages = []string{}

	// 统计服务器问题
	if report.ServerReport != nil {
		for _, issue := range report.ServerReport.Issues {
			summary.TotalIssues++
			switch issue.Level {
			case "critical":
				summary.CriticalIssues++
			case "warning":
				summary.WarningIssues++
			case "info":
				summary.InfoIssues++
			}

			// 添加关键和警告问题到消息列表
			if issue.Level == "critical" || issue.Level == "warning" {
				summary.Messages = append(summary.Messages, 
					fmt.Sprintf("[%s/%s] %s", report.ServerReport.Host, issue.Category, issue.Message))
			}
		}
	}

	// 统计K8s问题
	if report.K8sReport != nil {
		for _, issue := range report.K8sReport.Issues {
			summary.TotalIssues++
			switch issue.Level {
			case "critical":
				summary.CriticalIssues++
			case "warning":
				summary.WarningIssues++
			case "info":
				summary.InfoIssues++
			}

			if issue.Level == "critical" || issue.Level == "warning" {
				summary.Messages = append(summary.Messages, 
					fmt.Sprintf("[k8s/%s] %s", issue.Category, issue.Message))
			}
		}
	}

	// 确定整体状态
	if summary.CriticalIssues > 0 {
		summary.Status = "critical"
	} else if summary.WarningIssues > 0 {
		summary.Status = "warning"
	} else {
		summary.Status = "healthy"
	}
}

// FormatDuration 格式化持续时间
func FormatDuration(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// FormatBytes 格式化字节数
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatBytesPerSecond 格式化每秒字节数
func FormatBytesPerSecond(bytesPerSec int64) string {
	return FormatBytes(bytesPerSec) + "/s"
}

// CalculatePercentage 计算百分比
func CalculatePercentage(used, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

// GetSeverityColor 获取严重级别颜色(用于终端输出)
func GetSeverityColor(level string) string {
	switch level {
	case "critical":
		return "\033[31m" // 红色
	case "warning":
		return "\033[33m" // 黄色
	case "info":
		return "\033[36m" // 青色
	default:
		return "\033[0m" // 默认
	}
}

// ResetColor 重置颜色
func ResetColor() string {
	return "\033[0m"
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ValidateConfig 验证配置
func ValidateConfig(host, user, password string, port int) error {
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if user == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	return nil
}
