// Package test contains high-level integration tests for the kube-agents platform.
//
// These tests require a running Kubernetes cluster. By default, the framework
// creates an ephemeral kind cluster. Set KUBECONFIG to use an existing cluster.
//
// Run with: go test -v -tags=integration ./test/ -timeout 10m
package test

import (
	"os"
	"testing"

	"github.com/kube-agents/kube-agents-test/internal/testutil"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func newFramework(t *testing.T) *testutil.Framework {
	t.Helper()

	opts := testutil.Options{
		AgentConfigs: []agent.AgentConfig{
			{Name: "scaling-agent", Image: "ghcr.io/kube-agents/scaling-agent:latest", Namespace: "kube-agents"},
			{Name: "quota-agent", Image: "ghcr.io/kube-agents/quota-agent:latest", Namespace: "kube-agents"},
		},
	}

	// Use an existing cluster if KUBECONFIG is set.
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		opts.ClusterProvider = &cluster.ExistingProvider{KubeconfigPath: kc}
	}

	return testutil.Setup(t, opts)
}

// TestScenarioLoading verifies that scenario YAML files parse correctly.
// This test does not need a cluster — it only validates the loader.
func TestScenarioLoading(t *testing.T) {
	scenarios, err := scenario.LoadDir("scenarios")
	if err != nil {
		t.Fatalf("failed to load scenarios: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatal("no scenarios found in scenarios/")
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			if s.Name == "" {
				t.Error("scenario name is empty")
			}
			if len(s.Agents) == 0 {
				t.Error("scenario has no agents")
			}
			t.Logf("loaded scenario %q with %d agents, timeout %s",
				s.Name, len(s.Agents), s.Timeout.Duration)
		})
	}
}

// TestScalingRespectsQuota is an example integration test.
// It runs the scaling-respects-quota scenario against a real cluster.
//
// Skip this test in CI if no cluster is available or kind is not installed.
func TestScalingRespectsQuota(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := newFramework(t)
	f.RunScenario(t, "scenarios/scaling-respects-quota.yaml")
}

// TestAllScenarios runs every scenario file in the scenarios/ directory.
func TestAllScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	f := newFramework(t)
	f.RunScenarioDir(t, "scenarios")
}
