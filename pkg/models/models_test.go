package models

import (
	"testing"
	"time"
)

func TestInspectionReport(t *testing.T) {
	report := &InspectionReport{
		Timestamp: time.Now(),
		Type:      "server",
		Summary: InspectionSummary{
			TotalIssues:    0,
			CriticalIssues: 0,
			WarningIssues:  0,
			InfoIssues:     0,
			Status:         "healthy",
			Messages:       []string{},
		},
	}

	if report.Type != "server" {
		t.Errorf("Expected type 'server', got '%s'", report.Type)
	}

	if report.Summary.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", report.Summary.Status)
	}
}

func TestServerReport(t *testing.T) {
	report := &ServerReport{
		Host:      "test-server",
		Timestamp: time.Now(),
		Issues:    []Issue{},
	}

	if report.Host != "test-server" {
		t.Errorf("Expected host 'test-server', got '%s'", report.Host)
	}

	if len(report.Issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(report.Issues))
	}
}

func TestK8sReport(t *testing.T) {
	report := &K8sReport{
		Timestamp: time.Now(),
		Issues:    []Issue{},
		ClusterInfo: ClusterInfo{
			Version:        "v1.28.0",
			NodeCount:      3,
			PodCount:       50,
			NamespaceCount: 10,
		},
	}

	if report.ClusterInfo.NodeCount != 3 {
		t.Errorf("Expected 3 nodes, got %d", report.ClusterInfo.NodeCount)
	}

	if report.ClusterInfo.Version != "v1.28.0" {
		t.Errorf("Expected version 'v1.28.0', got '%s'", report.ClusterInfo.Version)
	}
}

func TestIssue(t *testing.T) {
	issue := Issue{
		Level:      "critical",
		Category:   "cpu",
		Message:    "High CPU usage",
		Details:    "CPU usage is above 90%",
		Timestamp:  time.Now(),
		Suggestion: "Check running processes",
	}

	if issue.Level != "critical" {
		t.Errorf("Expected level 'critical', got '%s'", issue.Level)
	}

	if issue.Category != "cpu" {
		t.Errorf("Expected category 'cpu', got '%s'", issue.Category)
	}
}

func TestCPUMetrics(t *testing.T) {
	metrics := CPUMetrics{
		CoreCount:     8,
		Load1:         2.5,
		Load5:         2.0,
		Load15:        1.5,
		UsagePercent:  75.5,
		IowaitPercent: 5.2,
	}

	if metrics.CoreCount != 8 {
		t.Errorf("Expected 8 cores, got %d", metrics.CoreCount)
	}

	if metrics.UsagePercent != 75.5 {
		t.Errorf("Expected usage 75.5%%, got %.2f%%", metrics.UsagePercent)
	}
}

func TestMemoryMetrics(t *testing.T) {
	metrics := MemoryMetrics{
		TotalMB:      16384,
		UsedMB:       8192,
		FreeMB:       4096,
		AvailableMB:  8192,
		UsagePercent: 50.0,
	}

	if metrics.TotalMB != 16384 {
		t.Errorf("Expected 16384 MB total, got %d", metrics.TotalMB)
	}

	if metrics.UsagePercent != 50.0 {
		t.Errorf("Expected 50%% usage, got %.2f%%", metrics.UsagePercent)
	}
}
