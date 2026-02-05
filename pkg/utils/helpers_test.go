package utils

import (
	"inspection-tool/pkg/models"
	"testing"
	"time"
)

func TestBuildInspectionSummary(t *testing.T) {
	report := &models.InspectionReport{
		Timestamp: time.Now(),
		Type:      "server",
		ServerReport: &models.ServerReport{
			Host: "test-server",
			Issues: []models.Issue{
				{Level: "critical", Category: "cpu", Message: "High CPU"},
				{Level: "warning", Category: "memory", Message: "High memory"},
				{Level: "info", Category: "disk", Message: "Info"},
			},
		},
	}

	BuildInspectionSummary(report)

	if report.Summary.TotalIssues != 3 {
		t.Errorf("Expected 3 total issues, got %d", report.Summary.TotalIssues)
	}

	if report.Summary.CriticalIssues != 1 {
		t.Errorf("Expected 1 critical issue, got %d", report.Summary.CriticalIssues)
	}

	if report.Summary.WarningIssues != 1 {
		t.Errorf("Expected 1 warning issue, got %d", report.Summary.WarningIssues)
	}

	if report.Summary.InfoIssues != 1 {
		t.Errorf("Expected 1 info issue, got %d", report.Summary.InfoIssues)
	}

	if report.Summary.Status != "critical" {
		t.Errorf("Expected status 'critical', got '%s'", report.Summary.Status)
	}

	if len(report.Summary.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(report.Summary.Messages))
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{30, "0m"},
		{90, "1m"},
		{3600, "1h 0m"},
		{3661, "1h 1m"},
		{86400, "1d 0h 0m"},
		{90061, "1d 1h 1m"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.seconds)
		if result != tt.expected {
			t.Errorf("FormatDuration(%d) = %s, expected %s", tt.seconds, result, tt.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestCalculatePercentage(t *testing.T) {
	tests := []struct {
		used     int64
		total    int64
		expected float64
	}{
		{50, 100, 50.0},
		{75, 100, 75.0},
		{0, 100, 0.0},
		{100, 100, 100.0},
		{50, 0, 0.0}, // Division by zero case
	}

	for _, tt := range tests {
		result := CalculatePercentage(tt.used, tt.total)
		if result != tt.expected {
			t.Errorf("CalculatePercentage(%d, %d) = %.2f, expected %.2f",
				tt.used, tt.total, result, tt.expected)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		host     string
		user     string
		password string
		port     int
		hasError bool
	}{
		{"192.168.1.100", "root", "pass", 22, false},
		{"", "root", "pass", 22, true},           // Empty host
		{"host", "", "pass", 22, true},           // Empty user
		{"host", "root", "pass", 0, true},        // Invalid port
		{"host", "root", "pass", 65536, true},    // Port too large
		{"host", "root", "pass", -1, true},       // Negative port
		{"example.com", "admin", "secret", 2222, false},
	}

	for _, tt := range tests {
		err := ValidateConfig(tt.host, tt.user, tt.password, tt.port)
		if (err != nil) != tt.hasError {
			t.Errorf("ValidateConfig(%s, %s, _, %d) error = %v, expected error = %v",
				tt.host, tt.user, tt.port, err, tt.hasError)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact", 5, "exact"},
		{"toolong", 5, "to..."},
	}

	for _, tt := range tests {
		result := TruncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("TruncateString(%s, %d) = %s, expected %s",
				tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestGetSeverityColor(t *testing.T) {
	tests := []struct {
		level    string
		expected string
	}{
		{"critical", "\033[31m"},
		{"warning", "\033[33m"},
		{"info", "\033[36m"},
		{"unknown", "\033[0m"},
	}

	for _, tt := range tests {
		result := GetSeverityColor(tt.level)
		if result != tt.expected {
			t.Errorf("GetSeverityColor(%s) = %s, expected %s",
				tt.level, result, tt.expected)
		}
	}
}

func TestResetColor(t *testing.T) {
	expected := "\033[0m"
	result := ResetColor()
	if result != expected {
		t.Errorf("ResetColor() = %s, expected %s", result, expected)
	}
}
