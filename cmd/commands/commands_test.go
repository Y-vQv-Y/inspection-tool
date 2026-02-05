package commands

import (
	"testing"
)

func TestServerOptions(t *testing.T) {
	opts := &ServerOptions{
		Host:     "192.168.1.100",
		User:     "root",
		Password: "test",
		Port:     22,
		Output:   "./reports",
		Format:   "json",
		Detailed: true,
	}

	if opts.Host != "192.168.1.100" {
		t.Errorf("Expected host '192.168.1.100', got '%s'", opts.Host)
	}

	if opts.Port != 22 {
		t.Errorf("Expected port 22, got %d", opts.Port)
	}

	if opts.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", opts.Format)
	}

	if !opts.Detailed {
		t.Error("Expected detailed to be true")
	}
}

func TestK8sOptions(t *testing.T) {
	opts := &K8sOptions{
		Kubeconfig:     "~/.kube/config",
		Namespaces:     "default,kube-system",
		Output:         "./reports",
		Format:         "yaml",
		Detailed:       true,
		InspectWorkers: false,
	}

	if opts.Kubeconfig != "~/.kube/config" {
		t.Errorf("Expected kubeconfig '~/.kube/config', got '%s'", opts.Kubeconfig)
	}

	if opts.Namespaces != "default,kube-system" {
		t.Errorf("Expected namespaces 'default,kube-system', got '%s'", opts.Namespaces)
	}

	if opts.InspectWorkers {
		t.Error("Expected InspectWorkers to be false")
	}
}

func TestAllOptions(t *testing.T) {
	opts := &AllOptions{
		Kubeconfig:  "~/.kube/config",
		Namespaces:  "default",
		Hosts:       "192.168.1.1,192.168.1.2",
		SSHUser:     "root",
		SSHPassword: "test",
		SSHPort:     22,
		Output:      "./reports",
		Format:      "json",
		Detailed:    true,
	}

	if opts.Kubeconfig != "~/.kube/config" {
		t.Errorf("Expected kubeconfig '~/.kube/config', got '%s'", opts.Kubeconfig)
	}

	if opts.Hosts != "192.168.1.1,192.168.1.2" {
		t.Errorf("Expected hosts '192.168.1.1,192.168.1.2', got '%s'", opts.Hosts)
	}

	if opts.SSHPort != 22 {
		t.Errorf("Expected SSH port 22, got %d", opts.SSHPort)
	}
}

func TestNewServerCommand(t *testing.T) {
	cmd := NewServerCommand()

	if cmd.Use != "server" {
		t.Errorf("Expected command name 'server', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
}

func TestNewK8sCommand(t *testing.T) {
	cmd := NewK8sCommand()

	if cmd.Use != "k8s" {
		t.Errorf("Expected command name 'k8s', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
}

func TestNewAllCommand(t *testing.T) {
	cmd := NewAllCommand()

	if cmd.Use != "all" {
		t.Errorf("Expected command name 'all', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty short description")
	}
}
