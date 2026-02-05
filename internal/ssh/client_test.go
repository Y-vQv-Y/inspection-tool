package ssh

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	config := &Config{
		Host:     "192.168.1.100",
		Port:     22,
		User:     "root",
		Password: "test",
		Timeout:  30 * time.Second,
	}

	if config.Host != "192.168.1.100" {
		t.Errorf("Expected host '192.168.1.100', got '%s'", config.Host)
	}

	if config.Port != 22 {
		t.Errorf("Expected port 22, got %d", config.Port)
	}

	if config.User != "root" {
		t.Errorf("Expected user 'root', got '%s'", config.User)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		Host:     "test.example.com",
		Port:     22,
		User:     "admin",
		Password: "pass",
	}

	if config.Timeout != 0 {
		t.Errorf("Expected default timeout 0, got %v", config.Timeout)
	}
}

// 注意: 实际的SSH连接测试需要真实的SSH服务器
// 这里只测试配置结构
func TestNewClientConfig(t *testing.T) {
	config := &Config{
		Host:     "",
		Port:     22,
		User:     "root",
		Password: "test",
	}

	// 测试空主机应该失败
	_, err := NewClient(config)
	if err == nil {
		t.Error("Expected error with empty host")
	}
}
