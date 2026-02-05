package server

import (
	"testing"
)

func TestExtractOSFamily(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ubuntu 20.04 LTS", "ubuntu"},
		{"CentOS Linux 7", "centos"},
		{"Red Hat Enterprise Linux 8", "redhat"},
		{"Debian GNU/Linux 11", "debian"},
		{"Unknown OS", "linux"},
	}

	for _, tt := range tests {
		result := extractOSFamily(tt.input)
		if result != tt.expected {
			t.Errorf("extractOSFamily(%s) = %s, expected %s",
				tt.input, result, tt.expected)
		}
	}
}

func TestParseUptime(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"12345.67", 12345},
		{"0.5", 0},
		{"999999.99", 999999},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := parseUptime(tt.input)
		if result != tt.expected {
			t.Errorf("parseUptime(%s) = %d, expected %d",
				tt.input, result, tt.expected)
		}
	}
}

func TestExtractFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"12.34 some text", 12.34},
		{"5.0", 5.0},
		{"  99.99  ", 99.99},
		{"invalid", 0.0},
	}

	for _, tt := range tests {
		result := extractFloat(tt.input)
		if result != tt.expected {
			t.Errorf("extractFloat(%s) = %.2f, expected %.2f",
				tt.input, result, tt.expected)
		}
	}
}

func TestParseCPUMetrics(t *testing.T) {
	coreCount := "8"
	loadavg := "2.5 2.0 1.5 1/100 12345"
	cpuUsage := "%Cpu(s):  5.0 us,  2.1 sy,  0.0 ni, 92.5 id,  0.3 wa,  0.0 hi,  0.1 si,  0.0 st"
	vmstat := `ctxt 12345678
intr 9876543
procs_running 3
procs_blocked 1`

	metrics := parseCPUMetrics(coreCount, loadavg, cpuUsage, vmstat)

	if metrics.CoreCount != 8 {
		t.Errorf("Expected 8 cores, got %d", metrics.CoreCount)
	}

	if metrics.Load1 != 2.5 {
		t.Errorf("Expected load1 2.5, got %.2f", metrics.Load1)
	}

	if metrics.IdlePercent != 92.5 {
		t.Errorf("Expected idle 92.5%%, got %.2f%%", metrics.IdlePercent)
	}

	if metrics.RunQueue != 3 {
		t.Errorf("Expected run queue 3, got %d", metrics.RunQueue)
	}

	if metrics.BlockedTasks != 1 {
		t.Errorf("Expected blocked tasks 1, got %d", metrics.BlockedTasks)
	}
}

func TestParseMemoryMetrics(t *testing.T) {
	meminfo := `MemTotal:       16384000 kB
MemFree:         4096000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
Cached:          2048000 kB
SwapTotal:       4096000 kB
SwapFree:        3072000 kB
Dirty:            102400 kB`

	memPressure := "some avg10=0.50 avg60=0.30 avg300=0.10 total=1000000"

	metrics := parseMemoryMetrics(meminfo, memPressure)

	if metrics.TotalMB != 16000 {
		t.Errorf("Expected total 16000 MB, got %d", metrics.TotalMB)
	}

	if metrics.AvailableMB != 8000 {
		t.Errorf("Expected available 8000 MB, got %d", metrics.AvailableMB)
	}

	if metrics.Pressure != "some" {
		t.Errorf("Expected pressure 'some', got '%s'", metrics.Pressure)
	}

	if metrics.SwapTotalMB != 4000 {
		t.Errorf("Expected swap total 4000 MB, got %d", metrics.SwapTotalMB)
	}
}

func TestParseTCPStats(t *testing.T) {
	tcpStats := `    100 ESTAB
     50 TIME-WAIT
     10 SYN-SENT
      5 LISTEN
      2 CLOSE-WAIT`

	stats := parseTCPStats(tcpStats)

	if stats.Established != 100 {
		t.Errorf("Expected 100 established, got %d", stats.Established)
	}

	if stats.TimeWait != 50 {
		t.Errorf("Expected 50 time-wait, got %d", stats.TimeWait)
	}

	if stats.SynSent != 10 {
		t.Errorf("Expected 10 syn-sent, got %d", stats.SynSent)
	}

	if stats.Listen != 5 {
		t.Errorf("Expected 5 listen, got %d", stats.Listen)
	}
}

func TestParseSystemMetrics(t *testing.T) {
	fileHandle := "1024	0	65536"
	procCount := "150"
	threadCount := "500"
	ntpStatus := "NTP synchronized: yes"
	timeOffset := "0.5"
	kernelParams := `net.core.somaxconn=128
net.ipv4.tcp_max_syn_backlog=512
fs.file-max=65536
vm.swappiness=60`

	metrics := parseSystemMetrics(fileHandle, procCount, threadCount, ntpStatus, timeOffset, kernelParams)

	if metrics.FileHandlesAllocated != 1024 {
		t.Errorf("Expected 1024 file handles, got %d", metrics.FileHandlesAllocated)
	}

	if metrics.FileHandlesMax != 65536 {
		t.Errorf("Expected max 65536, got %d", metrics.FileHandlesMax)
	}

	if !metrics.NTPSynced {
		t.Error("Expected NTP to be synced")
	}

	if metrics.ProcessCount != 149 {
		t.Errorf("Expected 149 processes, got %d", metrics.ProcessCount)
	}

	if len(metrics.KernelParams) != 4 {
		t.Errorf("Expected 4 kernel params, got %d", len(metrics.KernelParams))
	}
}
