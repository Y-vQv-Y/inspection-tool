package report

import (
	"inspection-tool/pkg/models"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator("json", "./test-reports", true)
	
	if gen.format != "json" {
		t.Errorf("Expected format 'json', got '%s'", gen.format)
	}
	
	if gen.outputDir != "./test-reports" {
		t.Errorf("Expected outputDir './test-reports', got '%s'", gen.outputDir)
	}
	
	if !gen.detailed {
		t.Error("Expected detailed to be true")
	}
}

func TestGenerateServerReport(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewGenerator("json", tmpDir, true)
	
	report := &models.ServerReport{
		Host:      "test-server",
		Timestamp: time.Now(),
		OS: models.OSInfo{
			Hostname: "test-host",
			Platform: "linux",
		},
		CPU: models.CPUMetrics{
			CoreCount: 8,
			Load1:     2.5,
		},
		Memory: models.MemoryMetrics{
			TotalMB: 16384,
		},
		Issues: []models.Issue{},
	}
	
	filePath, err := gen.GenerateServerReport(report)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Report file not created: %s", filePath)
	}
	
	// Cleanup is automatic with t.TempDir()
}

func TestGenerateK8sReport(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewGenerator("yaml", tmpDir, false)
	
	report := &models.K8sReport{
		Timestamp: time.Now(),
		ClusterInfo: models.ClusterInfo{
			Version:   "v1.28.0",
			NodeCount: 3,
		},
		Nodes:  []models.NodeMetrics{},
		Pods:   []models.PodMetrics{},
		Issues: []models.Issue{},
	}
	
	filePath, err := gen.GenerateK8sReport(report)
	if err != nil {
		t.Fatalf("Failed to generate K8s report: %v", err)
	}
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("K8s report file not created: %s", filePath)
	}
	
	if filepath.Ext(filePath) != ".yaml" {
		t.Errorf("Expected .yaml extension, got %s", filepath.Ext(filePath))
	}
}

func TestGenerateFullReport(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewGenerator("json", tmpDir, true)
	
	report := &models.InspectionReport{
		Timestamp: time.Now(),
		Type:      "all",
		Summary: models.InspectionSummary{
			TotalIssues:    5,
			CriticalIssues: 2,
			WarningIssues:  3,
			Status:         "warning",
		},
	}
	
	filePath, err := gen.GenerateFullReport(report)
	if err != nil {
		t.Fatalf("Failed to generate full report: %v", err)
	}
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Full report file not created: %s", filePath)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	gen := NewGenerator("xml", tmpDir, true)
	
	report := &models.ServerReport{
		Host:      "test",
		Timestamp: time.Now(),
	}
	
	_, err := gen.GenerateServerReport(report)
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
}

func TestGetHealthStatus(t *testing.T) {
	tests := []struct {
		healthy  bool
		expected string
	}{
		{true, "✓ Healthy"},
		{false, "✗ Unhealthy"},
	}
	
	for _, tt := range tests {
		result := getHealthStatus(tt.healthy)
		if result != tt.expected {
			t.Errorf("getHealthStatus(%v) = %s, expected %s", 
				tt.healthy, result, tt.expected)
		}
	}
}

func TestCleanupOldReports(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create some test files
	oldFile := filepath.Join(tmpDir, "old-report.json")
	newFile := filepath.Join(tmpDir, "new-report.json")
	
	if err := os.WriteFile(oldFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newFile, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Change modification time of old file
	oldTime := time.Now().AddDate(0, 0, -31)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	
	// Clean up files older than 30 days
	err := CleanupOldReports(tmpDir, 30)
	if err != nil {
		t.Fatalf("CleanupOldReports failed: %v", err)
	}
	
	// Old file should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been deleted")
	}
	
	// New file should still exist
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Error("New file should still exist")
	}
}
