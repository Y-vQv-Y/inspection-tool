package server

import (
	"inspection-tool/pkg/models"
	"math"
	"strconv"
	"strings"
)

// extractOSFamily 提取操作系统家族
func extractOSFamily(osRelease string) string {
	s := strings.ToLower(osRelease)

	switch {
	case strings.Contains(s, "red hat"),
		strings.Contains(s, "redhat"),
		strings.Contains(s, "rhel"):
		return "redhat"
	case strings.Contains(s, "centos"):
		return "centos"
	case strings.Contains(s, "ubuntu"):
		return "ubuntu"
	case strings.Contains(s, "debian"):
		return "debian"
	case strings.Contains(s, "linux"):
		return "linux"
	default:
		return "linux"
	}
}

// extractOSVersion 提取操作系统版本
func extractOSVersion(osRelease string) string {
	lines := strings.Split(osRelease, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VERSION=") || strings.HasPrefix(line, "VERSION_ID=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(parts[1], "\"")
			}
		}
	}
	return "unknown"
}

// parseUptime 解析运行时间
func parseUptime(uptime string) int64 {
	uptime = strings.TrimSpace(uptime)
	val, err := strconv.ParseFloat(uptime, 64)
	if err != nil {
		return 0
	}
	return int64(val)
}

// parseCPUMetrics 解析CPU指标
func parseCPUMetrics(coreCountStr, loadavg, cpuUsage, vmstat string) models.CPUMetrics {
	metrics := models.CPUMetrics{}
	
	// 核心数
	coreCount, _ := strconv.Atoi(strings.TrimSpace(coreCountStr))
	metrics.CoreCount = coreCount
	
	// 负载
	parts := strings.Fields(strings.TrimSpace(loadavg))
	if len(parts) >= 3 {
		metrics.Load1, _ = strconv.ParseFloat(parts[0], 64)
		metrics.Load5, _ = strconv.ParseFloat(parts[1], 64)
		metrics.Load15, _ = strconv.ParseFloat(parts[2], 64)
	}
	
	// CPU使用率 (从top命令解析)
	// 示例: %Cpu(s):  5.0 us,  2.1 sy,  0.0 ni, 92.5 id,  0.3 wa,  0.0 hi,  0.1 si,  0.0 st
	cpuUsage = strings.TrimSpace(cpuUsage)
	if strings.Contains(cpuUsage, "Cpu(s)") {
		parts := strings.Split(cpuUsage, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "us") {
				val := extractFloat(part)
				metrics.UserPercent = val
			} else if strings.Contains(part, "sy") {
				val := extractFloat(part)
				metrics.SystemPercent = val
			} else if strings.Contains(part, "id") {
				val := extractFloat(part)
				metrics.IdlePercent = val
			} else if strings.Contains(part, "wa") {
				val := extractFloat(part)
				metrics.IowaitPercent = val
			} else if strings.Contains(part, "st") {
				val := extractFloat(part)
				metrics.StealPercent = val
			}
		}
		metrics.UsagePercent = 100 - metrics.IdlePercent
	}
	
	// vmstat信息
	lines := strings.Split(strings.TrimSpace(vmstat), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ctxt") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				metrics.ContextSwitches, _ = strconv.ParseInt(parts[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "intr") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				metrics.Interrupts, _ = strconv.ParseInt(parts[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "procs_running") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				val, _ := strconv.Atoi(parts[1])
				metrics.RunQueue = val
			}
		} else if strings.HasPrefix(line, "procs_blocked") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				val, _ := strconv.Atoi(parts[1])
				metrics.BlockedTasks = val
			}
		}
	}
	
	return metrics
}

// parseMemoryMetrics 解析内存指标
func parseMemoryMetrics(meminfo, memPressure string) models.MemoryMetrics {
	metrics := models.MemoryMetrics{}
	
	// 解析meminfo
	meminfoMap := make(map[string]int64)
	lines := strings.Split(strings.TrimSpace(meminfo), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.ParseInt(parts[1], 10, 64)
			meminfoMap[key] = val
		}
	}
	
	// 转换为MB
	metrics.TotalMB = meminfoMap["MemTotal"] / 1024
	metrics.FreeMB = meminfoMap["MemFree"] / 1024
	metrics.AvailableMB = meminfoMap["MemAvailable"] / 1024
	metrics.CachedMB = meminfoMap["Cached"] / 1024
	metrics.BuffersMB = meminfoMap["Buffers"] / 1024
	metrics.SwapTotalMB = meminfoMap["SwapTotal"] / 1024
	metrics.SwapUsedMB = (meminfoMap["SwapTotal"] - meminfoMap["SwapFree"]) / 1024
	metrics.DirtyMB = meminfoMap["Dirty"] / 1024
	
	metrics.UsedMB = metrics.TotalMB - metrics.FreeMB - metrics.BuffersMB - metrics.CachedMB
	if metrics.TotalMB > 0 {
		metrics.UsagePercent = float64(metrics.TotalMB-metrics.AvailableMB) / float64(metrics.TotalMB) * 100
	}
	
	if metrics.SwapTotalMB > 0 {
		metrics.SwapPercent = float64(metrics.SwapUsedMB) / float64(metrics.SwapTotalMB) * 100
	}
	
	// 内存压力
	memPressure = strings.TrimSpace(memPressure)
	if strings.Contains(memPressure, "full") {
		metrics.Pressure = "full"
	} else if strings.Contains(memPressure, "some") {
		metrics.Pressure = "some"
	} else {
		metrics.Pressure = "none"
	}
	
	return metrics
}

// parseDiskMetrics 解析磁盘指标
func parseDiskMetrics(df, dfInodes, iostat1, iostat2, ioErrors string) []models.DiskMetrics {
	var disks []models.DiskMetrics
	
	// 解析df输出
	dfMap := make(map[string]models.DiskMetrics)
	lines := strings.Split(strings.TrimSpace(df), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // 跳过标题行
		}
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			disk := models.DiskMetrics{
				Device:     parts[0],
				MountPoint: parts[5],
				FsType:     parts[1],
			}
			
			// 解析大小 (以G为单位)
			totalStr := strings.TrimSuffix(parts[1], "G")
			usedStr := strings.TrimSuffix(parts[2], "G")
			freeStr := strings.TrimSuffix(parts[3], "G")
			
			disk.TotalGB, _ = strconv.ParseFloat(totalStr, 64)
			disk.UsedGB, _ = strconv.ParseFloat(usedStr, 64)
			disk.FreeGB, _ = strconv.ParseFloat(freeStr, 64)
			
			// 使用率
			usePercentStr := strings.TrimSuffix(parts[4], "%")
			disk.UsagePercent, _ = strconv.ParseFloat(usePercentStr, 64)
			
			dfMap[parts[5]] = disk
		}
	}
	
	// 解析inode
	lines = strings.Split(strings.TrimSpace(dfInodes), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			mountPoint := parts[5]
			if disk, ok := dfMap[mountPoint]; ok {
				disk.InodesTotal, _ = strconv.ParseInt(parts[1], 10, 64)
				disk.InodesUsed, _ = strconv.ParseInt(parts[2], 10, 64)
				disk.InodesFree, _ = strconv.ParseInt(parts[3], 10, 64)
				
				inodePercentStr := strings.TrimSuffix(parts[4], "%")
				disk.InodesPercent, _ = strconv.ParseFloat(inodePercentStr, 64)
				
				dfMap[mountPoint] = disk
			}
		}
	}
	
	// 解析IO统计 (简化版,实际需要计算差值)
	// 这里只做示例,实际需要解析/proc/diskstats并计算两次采样的差值
	iostat1Map := parseIOStats(iostat1)
	iostat2Map := parseIOStats(iostat2)
	
	for device, stats2 := range iostat2Map {
		if stats1, ok := iostat1Map[device]; ok {
			// 寻找对应的磁盘
			for mountPoint, disk := range dfMap {
				if strings.HasPrefix(disk.Device, "/dev/"+device) {
					// 计算每秒IO
					disk.ReadBytesPS = (stats2.ReadBytes - stats1.ReadBytes)
					disk.WriteBytesPS = (stats2.WriteBytes - stats1.WriteBytes)
					disk.ReadOpsPS = (stats2.ReadOps - stats1.ReadOps)
					disk.WriteOpsPS = (stats2.WriteOps - stats1.WriteOps)
					
					// 计算IO利用率 (简化计算)
					totalTime := stats2.IOTime - stats1.IOTime
					if totalTime > 0 {
						disk.IOUtilPercent = float64(totalTime) / 10.0 // 近似值
						if disk.IOUtilPercent > 100 {
							disk.IOUtilPercent = 100
						}
					}
					
					// 平均等待时间
					totalOps := disk.ReadOpsPS + disk.WriteOpsPS
					if totalOps > 0 {
						disk.AvgAwaitMs = float64(totalTime) / float64(totalOps)
					}
					
					dfMap[mountPoint] = disk
				}
			}
		}
	}
	
	// 转换为切片
	for _, disk := range dfMap {
		disks = append(disks, disk)
	}
	
	return disks
}

// IOStats IO统计
type IOStats struct {
	ReadOps    int64
	WriteOps   int64
	ReadBytes  int64
	WriteBytes int64
	IOTime     int64
}

// parseIOStats 解析/proc/diskstats
func parseIOStats(diskstats string) map[string]IOStats {
	stats := make(map[string]IOStats)
	
	lines := strings.Split(strings.TrimSpace(diskstats), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 14 {
			device := parts[2]
			
			// 跳过分区
			if strings.Contains(device, "loop") || len(device) > 10 {
				continue
			}
			
			stat := IOStats{}
			stat.ReadOps, _ = strconv.ParseInt(parts[3], 10, 64)
			stat.ReadBytes, _ = strconv.ParseInt(parts[5], 10, 64)
			stat.ReadBytes *= 512 // 扇区转字节
			
			stat.WriteOps, _ = strconv.ParseInt(parts[7], 10, 64)
			stat.WriteBytes, _ = strconv.ParseInt(parts[9], 10, 64)
			stat.WriteBytes *= 512
			
			stat.IOTime, _ = strconv.ParseInt(parts[12], 10, 64)
			
			stats[device] = stat
		}
	}
	
	return stats
}

// parseNetworkMetrics 解析网络指标
func parseNetworkMetrics(netdev1, netdev2, tcpStats, netstat, localIP string) models.NetworkMetrics {
	metrics := models.NetworkMetrics{
		Interfaces: []models.NetworkInterface{},
	}
	
	// 解析网络接口统计
	netdev1Map := parseNetDev(netdev1)
	netdev2Map := parseNetDev(netdev2)
	
	for name, stats2 := range netdev2Map {
		if name == "lo" {
			continue // 跳过回环接口
		}
		
		if stats1, ok := netdev1Map[name]; ok {
			iface := models.NetworkInterface{
				Name:        name,
				RxBytesPS:   stats2.RxBytes - stats1.RxBytes,
				TxBytesPS:   stats2.TxBytes - stats1.TxBytes,
				RxPacketsPS: stats2.RxPackets - stats1.RxPackets,
				TxPacketsPS: stats2.TxPackets - stats1.TxPackets,
				RxErrors:    stats2.RxErrors,
				TxErrors:    stats2.TxErrors,
				RxDropped:   stats2.RxDropped,
				TxDropped:   stats2.TxDropped,
			}
			
			// 计算错误率
			totalPackets := iface.RxPacketsPS + iface.TxPacketsPS
			if totalPackets > 0 {
				totalErrors := float64(iface.RxErrors + iface.TxErrors)
				iface.ErrorRate = totalErrors / float64(totalPackets)
			}
			
			metrics.Interfaces = append(metrics.Interfaces, iface)
			metrics.PacketErrors += iface.RxErrors + iface.TxErrors
			metrics.PacketDrops += iface.RxDropped + iface.TxDropped
		}
	}
	
	// 解析TCP连接统计
	metrics.TCPConnections = parseTCPStats(tcpStats)
	
	// 解析TCP重传
	if strings.Contains(netstat, "TcpExt") {
		lines := strings.Split(netstat, "\n")
		if len(lines) >= 2 {
			keys := strings.Fields(lines[0])
			values := strings.Fields(lines[1])
			
			for i, key := range keys {
				if i < len(values) {
					if key == "TCPRetrans" {
						val, _ := strconv.ParseInt(values[i], 10, 64)
						metrics.TCPConnections.Retransmits = val
					}
				}
			}
		}
	}
	
	return metrics
}

// NetDevStats 网络设备统计
type NetDevStats struct {
	RxBytes   int64
	RxPackets int64
	RxErrors  int64
	RxDropped int64
	TxBytes   int64
	TxPackets int64
	TxErrors  int64
	TxDropped int64
}

// parseNetDev 解析/proc/net/dev
func parseNetDev(netdev string) map[string]NetDevStats {
	stats := make(map[string]NetDevStats)
	
	lines := strings.Split(strings.TrimSpace(netdev), "\n")
	for i, line := range lines {
		if i <= 1 {
			continue // 跳过头部
		}
		
		parts := strings.Fields(line)
		if len(parts) >= 17 {
			name := strings.TrimSuffix(parts[0], ":")
			
			stat := NetDevStats{}
			stat.RxBytes, _ = strconv.ParseInt(parts[1], 10, 64)
			stat.RxPackets, _ = strconv.ParseInt(parts[2], 10, 64)
			stat.RxErrors, _ = strconv.ParseInt(parts[3], 10, 64)
			stat.RxDropped, _ = strconv.ParseInt(parts[4], 10, 64)
			
			stat.TxBytes, _ = strconv.ParseInt(parts[9], 10, 64)
			stat.TxPackets, _ = strconv.ParseInt(parts[10], 10, 64)
			stat.TxErrors, _ = strconv.ParseInt(parts[11], 10, 64)
			stat.TxDropped, _ = strconv.ParseInt(parts[12], 10, 64)
			
			stats[name] = stat
		}
	}
	
	return stats
}

// parseTCPStats 解析TCP统计
func parseTCPStats(tcpStats string) models.TCPStats {
	stats := models.TCPStats{}
	
	lines := strings.Split(strings.TrimSpace(tcpStats), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			count, _ := strconv.Atoi(parts[0])
			state := parts[1]
			
			switch state {
			case "ESTAB":
				stats.Established = count
			case "SYN-SENT":
				stats.SynSent = count
			case "SYN-RECV":
				stats.SynRecv = count
			case "FIN-WAIT-1":
				stats.FinWait1 = count
			case "FIN-WAIT-2":
				stats.FinWait2 = count
			case "TIME-WAIT":
				stats.TimeWait = count
			case "CLOSE-WAIT":
				stats.CloseWait = count
			case "LAST-ACK":
				stats.LastAck = count
			case "LISTEN":
				stats.Listen = count
			case "CLOSING":
				stats.Closing = count
			}
		}
	}
	
	return stats
}

// parseSystemMetrics 解析系统指标
func parseSystemMetrics(fileHandle, procCount, threadCount, ntpStatus, timeOffset, kernelParams string) models.SystemMetrics {
	metrics := models.SystemMetrics{
		KernelParams: make(map[string]string),
	}
	
	// 文件句柄
	parts := strings.Fields(strings.TrimSpace(fileHandle))
	if len(parts) >= 3 {
		metrics.FileHandlesAllocated, _ = strconv.ParseInt(parts[0], 10, 64)
		metrics.FileHandlesMax, _ = strconv.ParseInt(parts[2], 10, 64)
		
		if metrics.FileHandlesMax > 0 {
			metrics.FileHandlesPercent = float64(metrics.FileHandlesAllocated) / float64(metrics.FileHandlesMax) * 100
		}
	}
	
	// 进程和线程数
	count, _ := strconv.Atoi(strings.TrimSpace(procCount))
	metrics.ProcessCount = count - 1 // 减去标题行
	
	count, _ = strconv.Atoi(strings.TrimSpace(threadCount))
	metrics.ThreadCount = count - 1
	
	// NTP同步状态
	metrics.NTPSynced = strings.Contains(ntpStatus, "synchronized: yes") || 
		strings.Contains(ntpStatus, "NTP synchronized: yes")
	
	// 时间偏差
	offset := strings.TrimSpace(timeOffset)
	metrics.TimeOffset, _ = strconv.ParseFloat(offset, 64)
	if math.IsNaN(metrics.TimeOffset) {
		metrics.TimeOffset = 0
	}
	
	// 内核参数
	lines := strings.Split(strings.TrimSpace(kernelParams), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			metrics.KernelParams[parts[0]] = parts[1]
		}
	}
	
	return metrics
}

// extractFloat 从字符串中提取浮点数
func extractFloat(s string) float64 {
	parts := strings.Fields(s)
	if len(parts) > 0 {
		val, _ := strconv.ParseFloat(parts[0], 64)
		return val
	}
	return 0
}
