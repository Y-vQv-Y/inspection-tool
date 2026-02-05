package models

import "time"

// InspectionReport 巡检报告
type InspectionReport struct {
	Timestamp    time.Time         `json:"timestamp" yaml:"timestamp"`
	Type         string            `json:"type" yaml:"type"` // server, k8s, all
	ServerReport *ServerReport     `json:"server_report,omitempty" yaml:"server_report,omitempty"`
	K8sReport    *K8sReport        `json:"k8s_report,omitempty" yaml:"k8s_report,omitempty"`
	Summary      InspectionSummary `json:"summary" yaml:"summary"`
}

// InspectionSummary 巡检摘要
type InspectionSummary struct {
	TotalIssues    int      `json:"total_issues" yaml:"total_issues"`
	CriticalIssues int      `json:"critical_issues" yaml:"critical_issues"`
	WarningIssues  int      `json:"warning_issues" yaml:"warning_issues"`
	InfoIssues     int      `json:"info_issues" yaml:"info_issues"`
	Status         string   `json:"status" yaml:"status"` // healthy, warning, critical
	Messages       []string `json:"messages" yaml:"messages"`
}

// ServerReport 服务器巡检报告
type ServerReport struct {
	Host      string               `json:"host" yaml:"host"`
	OS        OSInfo               `json:"os" yaml:"os"`
	CPU       CPUMetrics           `json:"cpu" yaml:"cpu"`
	Memory    MemoryMetrics        `json:"memory" yaml:"memory"`
	Disk      []DiskMetrics        `json:"disk" yaml:"disk"`
	Network   NetworkMetrics       `json:"network" yaml:"network"`
	System    SystemMetrics        `json:"system" yaml:"system"`
	Issues    []Issue              `json:"issues" yaml:"issues"`
	Timestamp time.Time            `json:"timestamp" yaml:"timestamp"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	Hostname  string `json:"hostname" yaml:"hostname"`
	Platform  string `json:"platform" yaml:"platform"`
	Family    string `json:"family" yaml:"family"`
	Version   string `json:"version" yaml:"version"`
	KernelVer string `json:"kernel_version" yaml:"kernel_version"`
	Uptime    int64  `json:"uptime" yaml:"uptime"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	CoreCount       int     `json:"core_count" yaml:"core_count"`
	Load1           float64 `json:"load_1min" yaml:"load_1min"`
	Load5           float64 `json:"load_5min" yaml:"load_5min"`
	Load15          float64 `json:"load_15min" yaml:"load_15min"`
	UsagePercent    float64 `json:"usage_percent" yaml:"usage_percent"`
	UserPercent     float64 `json:"user_percent" yaml:"user_percent"`
	SystemPercent   float64 `json:"system_percent" yaml:"system_percent"`
	IdlePercent     float64 `json:"idle_percent" yaml:"idle_percent"`
	IowaitPercent   float64 `json:"iowait_percent" yaml:"iowait_percent"`
	StealPercent    float64 `json:"steal_percent" yaml:"steal_percent"`
	ContextSwitches int64   `json:"context_switches" yaml:"context_switches"`
	Interrupts      int64   `json:"interrupts" yaml:"interrupts"`
	RunQueue        int     `json:"run_queue" yaml:"run_queue"`
	BlockedTasks    int     `json:"blocked_tasks" yaml:"blocked_tasks"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	TotalMB       int64   `json:"total_mb" yaml:"total_mb"`
	UsedMB        int64   `json:"used_mb" yaml:"used_mb"`
	FreeMB        int64   `json:"free_mb" yaml:"free_mb"`
	AvailableMB   int64   `json:"available_mb" yaml:"available_mb"`
	UsagePercent  float64 `json:"usage_percent" yaml:"usage_percent"`
	CachedMB      int64   `json:"cached_mb" yaml:"cached_mb"`
	BuffersMB     int64   `json:"buffers_mb" yaml:"buffers_mb"`
	SwapTotalMB   int64   `json:"swap_total_mb" yaml:"swap_total_mb"`
	SwapUsedMB    int64   `json:"swap_used_mb" yaml:"swap_used_mb"`
	SwapPercent   float64 `json:"swap_percent" yaml:"swap_percent"`
	DirtyMB       int64   `json:"dirty_mb" yaml:"dirty_mb"`
	Pressure      string  `json:"pressure" yaml:"pressure"` // none, some, full
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	Device         string  `json:"device" yaml:"device"`
	MountPoint     string  `json:"mount_point" yaml:"mount_point"`
	FsType         string  `json:"fs_type" yaml:"fs_type"`
	TotalGB        float64 `json:"total_gb" yaml:"total_gb"`
	UsedGB         float64 `json:"used_gb" yaml:"used_gb"`
	FreeGB         float64 `json:"free_gb" yaml:"free_gb"`
	UsagePercent   float64 `json:"usage_percent" yaml:"usage_percent"`
	InodesTotal    int64   `json:"inodes_total" yaml:"inodes_total"`
	InodesUsed     int64   `json:"inodes_used" yaml:"inodes_used"`
	InodesFree     int64   `json:"inodes_free" yaml:"inodes_free"`
	InodesPercent  float64 `json:"inodes_percent" yaml:"inodes_percent"`
	ReadBytesPS    int64   `json:"read_bytes_per_sec" yaml:"read_bytes_per_sec"`
	WriteBytesPS   int64   `json:"write_bytes_per_sec" yaml:"write_bytes_per_sec"`
	ReadOpsPS      int64   `json:"read_ops_per_sec" yaml:"read_ops_per_sec"`
	WriteOpsPS     int64   `json:"write_ops_per_sec" yaml:"write_ops_per_sec"`
	IOUtilPercent  float64 `json:"io_util_percent" yaml:"io_util_percent"`
	AvgQueueSize   float64 `json:"avg_queue_size" yaml:"avg_queue_size"`
	AvgAwaitMs     float64 `json:"avg_await_ms" yaml:"avg_await_ms"`
	IOErrors       int64   `json:"io_errors" yaml:"io_errors"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	Interfaces      []NetworkInterface `json:"interfaces" yaml:"interfaces"`
	TCPConnections  TCPStats           `json:"tcp_connections" yaml:"tcp_connections"`
	PacketErrors    int64              `json:"packet_errors" yaml:"packet_errors"`
	PacketDrops     int64              `json:"packet_drops" yaml:"packet_drops"`
}

// NetworkInterface 网络接口
type NetworkInterface struct {
	Name           string  `json:"name" yaml:"name"`
	RxBytesPS      int64   `json:"rx_bytes_per_sec" yaml:"rx_bytes_per_sec"`
	TxBytesPS      int64   `json:"tx_bytes_per_sec" yaml:"tx_bytes_per_sec"`
	RxPacketsPS    int64   `json:"rx_packets_per_sec" yaml:"rx_packets_per_sec"`
	TxPacketsPS    int64   `json:"tx_packets_per_sec" yaml:"tx_packets_per_sec"`
	RxErrors       int64   `json:"rx_errors" yaml:"rx_errors"`
	TxErrors       int64   `json:"tx_errors" yaml:"tx_errors"`
	RxDropped      int64   `json:"rx_dropped" yaml:"rx_dropped"`
	TxDropped      int64   `json:"tx_dropped" yaml:"tx_dropped"`
	ErrorRate      float64 `json:"error_rate" yaml:"error_rate"`
}

// TCPStats TCP统计
type TCPStats struct {
	Established    int   `json:"established" yaml:"established"`
	SynSent        int   `json:"syn_sent" yaml:"syn_sent"`
	SynRecv        int   `json:"syn_recv" yaml:"syn_recv"`
	FinWait1       int   `json:"fin_wait1" yaml:"fin_wait1"`
	FinWait2       int   `json:"fin_wait2" yaml:"fin_wait2"`
	TimeWait       int   `json:"time_wait" yaml:"time_wait"`
	CloseWait      int   `json:"close_wait" yaml:"close_wait"`
	LastAck        int   `json:"last_ack" yaml:"last_ack"`
	Listen         int   `json:"listen" yaml:"listen"`
	Closing        int   `json:"closing" yaml:"closing"`
	Retransmits    int64 `json:"retransmits" yaml:"retransmits"`
	RetransmitRate float64 `json:"retransmit_rate" yaml:"retransmit_rate"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	FileHandlesAllocated int64   `json:"file_handles_allocated" yaml:"file_handles_allocated"`
	FileHandlesMax       int64   `json:"file_handles_max" yaml:"file_handles_max"`
	FileHandlesPercent   float64 `json:"file_handles_percent" yaml:"file_handles_percent"`
	ProcessCount         int     `json:"process_count" yaml:"process_count"`
	ThreadCount          int     `json:"thread_count" yaml:"thread_count"`
	TimeOffset           float64 `json:"time_offset_seconds" yaml:"time_offset_seconds"`
	NTPSynced            bool    `json:"ntp_synced" yaml:"ntp_synced"`
	KernelParams         map[string]string `json:"kernel_params" yaml:"kernel_params"`
}

// Issue 问题项
type Issue struct {
	Level       string    `json:"level" yaml:"level"` // critical, warning, info
	Category    string    `json:"category" yaml:"category"`
	Message     string    `json:"message" yaml:"message"`
	Details     string    `json:"details" yaml:"details"`
	Timestamp   time.Time `json:"timestamp" yaml:"timestamp"`
	Suggestion  string    `json:"suggestion" yaml:"suggestion"`
}

// K8sReport Kubernetes巡检报告
type K8sReport struct {
	ClusterInfo      ClusterInfo        `json:"cluster_info" yaml:"cluster_info"`
	Nodes            []NodeMetrics      `json:"nodes" yaml:"nodes"`
	APIServerStatus  APIServerMetrics   `json:"apiserver_status" yaml:"apiserver_status"`
	EtcdStatus       EtcdMetrics        `json:"etcd_status" yaml:"etcd_status"`
	ControllerStatus ControllerMetrics  `json:"controller_status" yaml:"controller_status"`
	SchedulerStatus  SchedulerMetrics   `json:"scheduler_status" yaml:"scheduler_status"`
	Pods             []PodMetrics       `json:"pods" yaml:"pods"`
	Issues           []Issue            `json:"issues" yaml:"issues"`
	Timestamp        time.Time          `json:"timestamp" yaml:"timestamp"`
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	Version       string `json:"version" yaml:"version"`
	NodeCount     int    `json:"node_count" yaml:"node_count"`
	PodCount      int    `json:"pod_count" yaml:"pod_count"`
	NamespaceCount int   `json:"namespace_count" yaml:"namespace_count"`
}

// NodeMetrics 节点指标
type NodeMetrics struct {
	Name              string            `json:"name" yaml:"name"`
	Ready             bool              `json:"ready" yaml:"ready"`
	Conditions        []NodeCondition   `json:"conditions" yaml:"conditions"`
	CPUCapacity       string            `json:"cpu_capacity" yaml:"cpu_capacity"`
	MemoryCapacity    string            `json:"memory_capacity" yaml:"memory_capacity"`
	PodsCapacity      int               `json:"pods_capacity" yaml:"pods_capacity"`
	CPUUsage          string            `json:"cpu_usage" yaml:"cpu_usage"`
	MemoryUsage       string            `json:"memory_usage" yaml:"memory_usage"`
	CPUPercent        float64           `json:"cpu_percent" yaml:"cpu_percent"`
	MemoryPercent     float64           `json:"memory_percent" yaml:"memory_percent"`
	PodCount          int               `json:"pod_count" yaml:"pod_count"`
	PodPercent        float64           `json:"pod_percent" yaml:"pod_percent"`
	Labels            map[string]string `json:"labels" yaml:"labels"`
	Taints            []string          `json:"taints" yaml:"taints"`
	KernelVersion     string            `json:"kernel_version" yaml:"kernel_version"`
	OSImage           string            `json:"os_image" yaml:"os_image"`
	ContainerRuntime  string            `json:"container_runtime" yaml:"container_runtime"`
	KubeletVersion    string            `json:"kubelet_version" yaml:"kubelet_version"`
}

// NodeCondition 节点状态
type NodeCondition struct {
	Type    string `json:"type" yaml:"type"`
	Status  string `json:"status" yaml:"status"`
	Reason  string `json:"reason" yaml:"reason"`
	Message string `json:"message" yaml:"message"`
}

// APIServerMetrics API Server指标
type APIServerMetrics struct {
	Healthy       bool   `json:"healthy" yaml:"healthy"`
	Version       string `json:"version" yaml:"version"`
	RequestRate   int64  `json:"request_rate" yaml:"request_rate"`
	ErrorRate     float64 `json:"error_rate" yaml:"error_rate"`
	Latency       LatencyMetrics `json:"latency" yaml:"latency"`
}

// LatencyMetrics 延迟指标
type LatencyMetrics struct {
	P50 float64 `json:"p50" yaml:"p50"`
	P95 float64 `json:"p95" yaml:"p95"`
	P99 float64 `json:"p99" yaml:"p99"`
}

// EtcdMetrics etcd指标
type EtcdMetrics struct {
	Healthy        bool    `json:"healthy" yaml:"healthy"`
	ClusterSize    int     `json:"cluster_size" yaml:"cluster_size"`
	Leader         string  `json:"leader" yaml:"leader"`
	DBSize         int64   `json:"db_size_mb" yaml:"db_size_mb"`
	LeaderChanges  int     `json:"leader_changes" yaml:"leader_changes"`
	Members        []EtcdMember `json:"members" yaml:"members"`
}

// EtcdMember etcd成员
type EtcdMember struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	IsLeader bool `json:"is_leader" yaml:"is_leader"`
}

// ControllerMetrics Controller Manager指标
type ControllerMetrics struct {
	Healthy bool `json:"healthy" yaml:"healthy"`
	Leader  string `json:"leader" yaml:"leader"`
}

// SchedulerMetrics Scheduler指标
type SchedulerMetrics struct {
	Healthy bool `json:"healthy" yaml:"healthy"`
	Leader  string `json:"leader" yaml:"leader"`
}

// PodMetrics Pod指标
type PodMetrics struct {
	Name              string            `json:"name" yaml:"name"`
	Namespace         string            `json:"namespace" yaml:"namespace"`
	Phase             string            `json:"phase" yaml:"phase"`
	Ready             bool              `json:"ready" yaml:"ready"`
	RestartCount      int               `json:"restart_count" yaml:"restart_count"`
	Node              string            `json:"node" yaml:"node"`
	CPURequest        string            `json:"cpu_request" yaml:"cpu_request"`
	MemoryRequest     string            `json:"memory_request" yaml:"memory_request"`
	CPULimit          string            `json:"cpu_limit" yaml:"cpu_limit"`
	MemoryLimit       string            `json:"memory_limit" yaml:"memory_limit"`
	CPUUsage          string            `json:"cpu_usage" yaml:"cpu_usage"`
	MemoryUsage       string            `json:"memory_usage" yaml:"memory_usage"`
	Conditions        []PodCondition    `json:"conditions" yaml:"conditions"`
	Labels            map[string]string `json:"labels" yaml:"labels"`
	Age               int64             `json:"age_seconds" yaml:"age_seconds"`
}

// PodCondition Pod状态
type PodCondition struct {
	Type    string `json:"type" yaml:"type"`
	Status  string `json:"status" yaml:"status"`
	Reason  string `json:"reason" yaml:"reason"`
	Message string `json:"message" yaml:"message"`
}
