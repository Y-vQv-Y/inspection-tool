package k8s

import (
	"inspection-tool/pkg/models"
	"testing"
)

func TestInspectorConfig(t *testing.T) {
	config := &InspectorConfig{
		Kubeconfig: "/path/to/kubeconfig",
		Namespaces: []string{"default", "kube-system"},
	}

	if config.Kubeconfig != "/path/to/kubeconfig" {
		t.Errorf("Expected kubeconfig path, got '%s'", config.Kubeconfig)
	}

	if len(config.Namespaces) != 2 {
		t.Errorf("Expected 2 namespaces, got %d", len(config.Namespaces))
	}
}

func TestGetNodeConditionDetails(t *testing.T) {
	conditions := []models.NodeCondition{
		{Type: "Ready", Status: "False", Reason: "KubeletNotReady", Message: "kubelet is not ready"},
		{Type: "DiskPressure", Status: "True", Reason: "KubeletHasDiskPressure", Message: "disk pressure"},
	}

	details := getNodeConditionDetails(conditions)
	
	if details == "" {
		t.Error("Expected non-empty details")
	}
	
	if !contains(details, "Ready") {
		t.Error("Expected details to contain 'Ready'")
	}
}

func TestGetPodConditionDetails(t *testing.T) {
	conditions := []models.PodCondition{
		{Type: "Ready", Status: "False", Reason: "ContainersNotReady", Message: "containers not ready"},
		{Type: "Initialized", Status: "True", Reason: "PodInitialized", Message: "initialized"},
	}

	details := getPodConditionDetails(conditions)
	
	if !contains(details, "Ready") {
		t.Error("Expected details to contain 'Ready'")
	}
}

func TestCountHealthyEtcdMembers(t *testing.T) {
	members := []models.EtcdMember{
		{Name: "etcd-1", Status: "Running", IsLeader: true},
		{Name: "etcd-2", Status: "Running", IsLeader: false},
		{Name: "etcd-3", Status: "Pending", IsLeader: false},
	}

	count := countHealthyEtcdMembers(members)
	
	if count != 2 {
		t.Errorf("Expected 2 healthy members, got %d", count)
	}
}

// Helper function for tests
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(s == substr || len(s) >= len(substr) && s[:len(substr)] == substr ||
		 len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		 stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
