package server

import (
	"fmt"
	"inspection-tool/internal/ssh"
	"inspection-tool/pkg/models"
	"strings"
	"time"
)

// Inspector 服务器巡检器
type Inspector struct {
	sshClient *ssh.Client
	localIP   string
}

// NewInspector 创建巡检器
func NewInspector(sshClient *ssh.Client) (*Inspector, error) {
	localIP, err := ssh.GetLocalIP()
	if err != nil {
		localIP = ""
	}

	return &Inspector{
		sshClient: sshClient,
		localIP:   localIP,
	}, nil
}

// Inspect 执行巡检
func (i *Inspector) Inspect() (*models.ServerReport, error) {
	report := &models.ServerReport{
		Host:      i.sshClient.GetHost(),
		Timestamp: time.Now(),
		Issues:    []models.Issue{},
	}

	// 收集操作系统信息
	if err := i.collectOSInfo(report); err != nil {
		return nil, fmt.Errorf("failed to collect OS info: %w", err)
	}

	// 收集CPU指标
	if err := i.collectCPUMetrics(report); err != nil {
		return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	// 收集内存指标
	if err := i.collectMemoryMetrics(report); err != nil {
		return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	// 收集磁盘指标
	if err := i.collectDiskMetrics(report); err != nil {
		return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
	}

	// 收集网络指标
	if err := i.collectNetworkMetrics(report); err != nil {
		return nil, fmt.Errorf("failed to collect network metrics: %w", err)
	}

	// 收集系统指标
	if err := i.collectSystemMetrics(report); err != nil {
		return nil, fmt.Errorf("failed to collect system metrics: %w", err)
	}

	// 分析问题
	i.analyzeIssues(report)

	return report, nil
}

// collectOSInfo 收集操作系统信息
func (i *Inspector) collectOSInfo(report *models.ServerReport) error {
	// hostname
	hostname, err := i.sshClient.Execute("hostname")
	if err != nil {
		return err
	}

	// 系统信息
	osRelease, _ := i.sshClient.Execute("cat /etc/os-release 2>/dev/null || cat /etc/redhat-release 2>/dev/null")
	kernelVer, _ := i.sshClient.Execute("uname -r")
	uptime, _ := i.sshClient.Execute("cat /proc/uptime | awk '{print $1}'")

	report.OS = models.OSInfo{
		Hostname:  strings.TrimSpace(hostname),
		Platform:  "linux",
		Family:    extractOSFamily(osRelease),
		Version:   extractOSVersion(osRelease),
		KernelVer: strings.TrimSpace(kernelVer),
		Uptime:    parseUptime(uptime),
	}

	return nil
}

// collectCPUMetrics 收集CPU指标
func (i *Inspector) collectCPUMetrics(report *models.ServerReport) error {
	// CPU核心数
	coreCount, _ := i.sshClient.Execute("grep -c ^processor /proc/cpuinfo")
	
	// 负载
	loadavg, _ := i.sshClient.Execute("cat /proc/loadavg")
	
	// CPU使用率 (通过top命令获取)
	cpuUsage, _ := i.sshClient.Execute("top -bn2 -d 0.5 | grep 'Cpu(s)' | tail -n 1")
	
	// 上下文切换和中断
	vmstat, _ := i.sshClient.Execute("cat /proc/stat | grep -E '^(ctxt|intr|procs_running|procs_blocked)'")
	
	// 解析数据
	report.CPU = parseCPUMetrics(coreCount, loadavg, cpuUsage, vmstat)
	
	return nil
}

// collectMemoryMetrics 收集内存指标
func (i *Inspector) collectMemoryMetrics(report *models.ServerReport) error {
	// 内存信息
	meminfo, _ := i.sshClient.Execute("cat /proc/meminfo")
	
	// 内存压力 (PSI - Pressure Stall Information)
	memPressure, _ := i.sshClient.Execute("cat /proc/pressure/memory 2>/dev/null || echo 'none'")
	
	report.Memory = parseMemoryMetrics(meminfo, memPressure)
	
	return nil
}

// collectDiskMetrics 收集磁盘指标
func (i *Inspector) collectDiskMetrics(report *models.ServerReport) error {
	// 磁盘使用情况
	df, _ := i.sshClient.Execute("df -BG -x tmpfs -x devtmpfs")
	
	// Inode使用情况
	dfInodes, _ := i.sshClient.Execute("df -i -x tmpfs -x devtmpfs")
	
	// IO统计 (需要两次采样)
	iostat1, _ := i.sshClient.Execute("cat /proc/diskstats")
	time.Sleep(1 * time.Second)
	iostat2, _ := i.sshClient.Execute("cat /proc/diskstats")
	
	// IO错误
	ioErrors, _ := i.sshClient.Execute("grep -r '' /sys/block/*/stat 2>/dev/null | grep -v '0 0 0 0 0 0 0 0 0 0 0' || echo ''")
	
	report.Disk = parseDiskMetrics(df, dfInodes, iostat1, iostat2, ioErrors)
	
	return nil
}

// collectNetworkMetrics 收集网络指标
func (i *Inspector) collectNetworkMetrics(report *models.ServerReport) error {
	// 网络接口统计 (两次采样计算速率)
	netdev1, _ := i.sshClient.Execute("cat /proc/net/dev")
	time.Sleep(1 * time.Second)
	netdev2, _ := i.sshClient.Execute("cat /proc/net/dev")
	
	// TCP连接统计
	tcpStats, _ := i.sshClient.Execute("ss -tan state all | tail -n +2 | awk '{print $1}' | sort | uniq -c")
	
	// TCP重传
	netstat, _ := i.sshClient.Execute("cat /proc/net/netstat | grep TcpExt")
	
	report.Network = parseNetworkMetrics(netdev1, netdev2, tcpStats, netstat, i.localIP)
	
	return nil
}

// collectSystemMetrics 收集系统指标
func (i *Inspector) collectSystemMetrics(report *models.ServerReport) error {
	// 文件句柄
	fileHandle, _ := i.sshClient.Execute("cat /proc/sys/fs/file-nr")
	
	// 进程和线程数
	procCount, _ := i.sshClient.Execute("ps aux | wc -l")
	threadCount, _ := i.sshClient.Execute("ps -eLf | wc -l")
	
	// 时间同步
	ntpStatus, _ := i.sshClient.Execute("timedatectl status 2>/dev/null || echo 'unknown'")
	timeOffset, _ := i.sshClient.Execute("ntpq -p 2>/dev/null | tail -n 1 | awk '{print $9}' || echo '0'")
	
	// 关键内核参数
	kernelParams, _ := i.sshClient.Execute(`
		echo "net.core.somaxconn=$(cat /proc/sys/net/core/somaxconn)"
		echo "net.ipv4.tcp_max_syn_backlog=$(cat /proc/sys/net/ipv4/tcp_max_syn_backlog)"
		echo "fs.file-max=$(cat /proc/sys/fs/file-max)"
		echo "vm.swappiness=$(cat /proc/sys/vm/swappiness)"
	`)
	
	report.System = parseSystemMetrics(fileHandle, procCount, threadCount, ntpStatus, timeOffset, kernelParams)
	
	return nil
}

// analyzeIssues 分析问题
func (i *Inspector) analyzeIssues(report *models.ServerReport) {
	// CPU问题分析
	if report.CPU.Load1 > float64(report.CPU.CoreCount)*2 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "critical",
			Category: "cpu",
			Message:  fmt.Sprintf("CPU负载过高: %.2f (核心数: %d)", report.CPU.Load1, report.CPU.CoreCount),
			Details:  "1分钟平均负载超过核心数的2倍",
			Timestamp: time.Now(),
			Suggestion: "检查高CPU进程,考虑优化或扩容",
		})
	}
	
	if report.CPU.IowaitPercent > 30 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "warning",
			Category: "cpu",
			Message:  fmt.Sprintf("IO等待时间过高: %.2f%%", report.CPU.IowaitPercent),
			Details:  "CPU大量时间在等待IO操作",
			Timestamp: time.Now(),
			Suggestion: "检查磁盘IO性能,优化IO密集型操作",
		})
	}
	
	if report.CPU.BlockedTasks > 10 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "warning",
			Category: "cpu",
			Message:  fmt.Sprintf("阻塞任务数量过多: %d", report.CPU.BlockedTasks),
			Details:  "有大量任务处于不可中断睡眠状态",
			Timestamp: time.Now(),
			Suggestion: "检查IO子系统和锁竞争问题",
		})
	}
	
	// 内存问题分析
	if report.Memory.UsagePercent > 90 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "critical",
			Category: "memory",
			Message:  fmt.Sprintf("内存使用率过高: %.2f%%", report.Memory.UsagePercent),
			Details:  fmt.Sprintf("可用内存: %d MB", report.Memory.AvailableMB),
			Timestamp: time.Now(),
			Suggestion: "释放内存或增加物理内存",
		})
	}
	
	if report.Memory.SwapPercent > 50 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "warning",
			Category: "memory",
			Message:  fmt.Sprintf("Swap使用率过高: %.2f%%", report.Memory.SwapPercent),
			Details:  "系统在使用交换空间,可能影响性能",
			Timestamp: time.Now(),
			Suggestion: "检查内存泄漏,考虑增加物理内存",
		})
	}
	
	if strings.Contains(report.Memory.Pressure, "some") || strings.Contains(report.Memory.Pressure, "full") {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "critical",
			Category: "memory",
			Message:  "检测到内存压力",
			Details:  fmt.Sprintf("内存压力状态: %s", report.Memory.Pressure),
			Timestamp: time.Now(),
			Suggestion: "系统正在经历内存压力,需要立即处理",
		})
	}
	
	// 磁盘问题分析
	for _, disk := range report.Disk {
		if disk.UsagePercent > 85 {
			report.Issues = append(report.Issues, models.Issue{
				Level:    "critical",
				Category: "disk",
				Message:  fmt.Sprintf("磁盘空间不足: %s (%.2f%%)", disk.MountPoint, disk.UsagePercent),
				Details:  fmt.Sprintf("剩余空间: %.2f GB", disk.FreeGB),
				Timestamp: time.Now(),
				Suggestion: "清理磁盘空间或扩容",
			})
		}
		
		if disk.InodesPercent > 80 {
			report.Issues = append(report.Issues, models.Issue{
				Level:    "warning",
				Category: "disk",
				Message:  fmt.Sprintf("Inode使用率过高: %s (%.2f%%)", disk.MountPoint, disk.InodesPercent),
				Details:  fmt.Sprintf("剩余Inode: %d", disk.InodesFree),
				Timestamp: time.Now(),
				Suggestion: "删除不需要的小文件",
			})
		}
		
		if disk.IOUtilPercent > 80 {
			report.Issues = append(report.Issues, models.Issue{
				Level:    "warning",
				Category: "disk",
				Message:  fmt.Sprintf("磁盘IO利用率过高: %s (%.2f%%)", disk.Device, disk.IOUtilPercent),
				Details:  fmt.Sprintf("平均等待时间: %.2f ms", disk.AvgAwaitMs),
				Timestamp: time.Now(),
				Suggestion: "优化IO操作或升级存储",
			})
		}
		
		if disk.IOErrors > 0 {
			report.Issues = append(report.Issues, models.Issue{
				Level:    "critical",
				Category: "disk",
				Message:  fmt.Sprintf("检测到磁盘IO错误: %s", disk.Device),
				Details:  fmt.Sprintf("错误计数: %d", disk.IOErrors),
				Timestamp: time.Now(),
				Suggestion: "检查磁盘健康状态,可能需要更换磁盘",
			})
		}
	}
	
	// 网络问题分析
	for _, iface := range report.Network.Interfaces {
		if iface.ErrorRate > 0.01 {
			report.Issues = append(report.Issues, models.Issue{
				Level:    "warning",
				Category: "network",
				Message:  fmt.Sprintf("网络接口错误率过高: %s (%.4f%%)", iface.Name, iface.ErrorRate*100),
				Details:  fmt.Sprintf("接收错误: %d, 发送错误: %d", iface.RxErrors, iface.TxErrors),
				Timestamp: time.Now(),
				Suggestion: "检查网络硬件和线缆",
			})
		}
	}
	
	if report.Network.TCPConnections.RetransmitRate > 0.05 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "warning",
			Category: "network",
			Message:  fmt.Sprintf("TCP重传率过高: %.2f%%", report.Network.TCPConnections.RetransmitRate*100),
			Details:  fmt.Sprintf("重传次数: %d", report.Network.TCPConnections.Retransmits),
			Timestamp: time.Now(),
			Suggestion: "检查网络质量和TCP参数配置",
		})
	}
	
	if report.Network.TCPConnections.TimeWait > 10000 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "info",
			Category: "network",
			Message:  fmt.Sprintf("TIME_WAIT连接数过多: %d", report.Network.TCPConnections.TimeWait),
			Details:  "可能影响可用端口数",
			Timestamp: time.Now(),
			Suggestion: "调整net.ipv4.tcp_tw_reuse参数",
		})
	}
	
	// 系统问题分析
	if report.System.FileHandlesPercent > 80 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "warning",
			Category: "system",
			Message:  fmt.Sprintf("文件句柄使用率过高: %.2f%%", report.System.FileHandlesPercent),
			Details:  fmt.Sprintf("已分配: %d, 最大值: %d", report.System.FileHandlesAllocated, report.System.FileHandlesMax),
			Timestamp: time.Now(),
			Suggestion: "增加fs.file-max参数或排查句柄泄漏",
		})
	}
	
	if report.System.TimeOffset > 5 || report.System.TimeOffset < -5 {
		report.Issues = append(report.Issues, models.Issue{
			Level:    "warning",
			Category: "system",
			Message:  fmt.Sprintf("时间偏差过大: %.2f秒", report.System.TimeOffset),
			Details:  "系统时间与NTP服务器不同步",
			Timestamp: time.Now(),
			Suggestion: "配置NTP服务并同步时间",
		})
	}
}

// Close 关闭巡检器
func (i *Inspector) Close() error {
	if i.sshClient != nil {
		return i.sshClient.Close()
	}
	return nil
}
