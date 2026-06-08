// Configure RunOptions for an attached (existing) cluster with sandbox constraints.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kube-agents/kube-agents-test/internal/stub"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	r := runner.NewDefault(runner.Dependencies{
		Provider: stub.NewProvider(),
		Manager:  stub.NewManager(),
		Engine:   stub.NewEngine(),
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{"examples/scenarios/reconcile-only/scenario.yaml"},
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeAttached,
			Attached: &cluster.AttachedConfig{
				KubeconfigPath: kubeconfig,
				Context:        os.Getenv("KUBE_CONTEXT"),
				LeaveRunning:   true,
			},
		},
		AgentConfig: agent.Config{
			Mode:      agent.DeployPods,
			Namespace: "test-agents",
			Images: map[string]string{
				"scaling-agent": "ghcr.io/example/scaling-agent:latest",
				"quota-agent":   "ghcr.io/example/quota-agent:latest",
			},
		},
		SandboxNamespace:   "test-agents",
		AllowClusterScoped: false,
		AllowFaults:        false,
		LeaveCluster:       true,
		ArtifactsDir:       "/tmp/kube-agents-test-artifacts",
	})

	// Stub provider always fails at Ensure — this example documents configuration only.
	if err != nil {
		fmt.Printf("attached run config applied; Ensure failed as expected with stubs: %v\n", err)
		os.Exit(0)
	}
}
