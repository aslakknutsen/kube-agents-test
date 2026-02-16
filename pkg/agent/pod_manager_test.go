package agent

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodManager_DeployAndStop(t *testing.T) {
	client := fake.NewSimpleClientset()

	configs := []AgentConfig{
		{Name: "test-agent", Image: "test:latest", Namespace: "agents"},
	}
	pm := NewPodManagerWithClient(client, configs)

	ctx := context.Background()

	// Deploy.
	if err := pm.Deploy(ctx, "test-agent"); err != nil {
		t.Fatalf("Deploy() error: %v", err)
	}

	// Verify deployment was created.
	dep, err := client.AppsV1().Deployments("agents").Get(ctx, "test-agent", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting deployment: %v", err)
	}
	if dep.Spec.Template.Spec.Containers[0].Image != "test:latest" {
		t.Errorf("image = %q, want %q", dep.Spec.Template.Spec.Containers[0].Image, "test:latest")
	}
	if *dep.Spec.Replicas != 1 {
		t.Errorf("replicas = %d, want 1", *dep.Spec.Replicas)
	}

	// Stop.
	if err := pm.Stop(ctx, "test-agent"); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	// Verify deployment was deleted.
	_, err = client.AppsV1().Deployments("agents").Get(ctx, "test-agent", metav1.GetOptions{})
	if err == nil {
		t.Error("expected deployment to be deleted, but it still exists")
	}
}

func TestPodManager_DeployUnknownAgent(t *testing.T) {
	client := fake.NewSimpleClientset()
	pm := NewPodManagerWithClient(client, nil)

	err := pm.Deploy(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestPodManager_StopNotDeployed(t *testing.T) {
	client := fake.NewSimpleClientset()
	configs := []AgentConfig{
		{Name: "test-agent", Image: "test:latest", Namespace: "default"},
	}
	pm := NewPodManagerWithClient(client, configs)

	err := pm.Stop(context.Background(), "test-agent")
	if err == nil {
		t.Fatal("expected error when stopping agent that isn't deployed")
	}
}

func TestPodManager_StopAll(t *testing.T) {
	client := fake.NewSimpleClientset()
	configs := []AgentConfig{
		{Name: "agent-a", Image: "a:latest", Namespace: "ns"},
		{Name: "agent-b", Image: "b:latest", Namespace: "ns"},
	}
	pm := NewPodManagerWithClient(client, configs)

	ctx := context.Background()
	pm.Deploy(ctx, "agent-a")
	pm.Deploy(ctx, "agent-b")

	if err := pm.StopAll(ctx); err != nil {
		t.Fatalf("StopAll() error: %v", err)
	}

	deps, _ := client.AppsV1().Deployments("ns").List(ctx, metav1.ListOptions{})
	if len(deps.Items) != 0 {
		t.Errorf("expected 0 deployments after StopAll, got %d", len(deps.Items))
	}
}

func TestPodManager_DefaultNamespace(t *testing.T) {
	client := fake.NewSimpleClientset()
	configs := []AgentConfig{
		{Name: "agent", Image: "img:latest"},
	}
	pm := NewPodManagerWithClient(client, configs)

	ctx := context.Background()
	if err := pm.Deploy(ctx, "agent"); err != nil {
		t.Fatalf("Deploy() error: %v", err)
	}

	_, err := client.AppsV1().Deployments("default").Get(ctx, "agent", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected deployment in 'default' namespace: %v", err)
	}
}
