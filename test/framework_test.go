package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// TestScenarioLoading verifies that scenario YAML files can be parsed.
func TestScenarioLoading(t *testing.T) {
	scenarioDir := filepath.Join("scenarios")

	scenarios, err := scenario.LoadDir(scenarioDir)
	if err != nil {
		t.Fatalf("failed to load scenarios: %v", err)
	}

	if len(scenarios) == 0 {
		t.Fatal("expected at least one scenario file")
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			if s.Name == "" {
				t.Error("scenario has empty name")
			}
			if len(s.Agents) == 0 {
				t.Error("scenario has no agents defined")
			}
			if s.Expect.Timeout.Duration == 0 {
				t.Error("scenario has no timeout")
			}
			t.Logf("loaded scenario %q with %d agents, timeout %s",
				s.Name, len(s.Agents), s.Expect.Timeout.Duration)
		})
	}
}

// TestScenarioParsing verifies parsing of inline scenario YAML.
func TestScenarioParsing(t *testing.T) {
	raw := []byte(`
name: test-scenario
description: a simple test
agents:
  - agent-a
  - agent-b
setup:
  manifests:
    - fixtures/ns.yaml
expect:
  resources:
    - apiVersion: v1
      kind: ConfigMap
      name: result
      namespace: test
      conditions:
        - path: .data.status
          value: done
  timeout: 30s
`)
	s, err := scenario.Parse(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if s.Name != "test-scenario" {
		t.Errorf("expected name 'test-scenario', got %q", s.Name)
	}
	if len(s.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(s.Agents))
	}
	if len(s.Expect.Resources) != 1 {
		t.Errorf("expected 1 resource expectation, got %d", len(s.Expect.Resources))
	}
	if s.Expect.Timeout.Seconds() != 30 {
		t.Errorf("expected 30s timeout, got %s", s.Expect.Timeout.Duration)
	}
}

// TestEngineLifecycle verifies the engine can be constructed and wired together.
// This is a smoke test for the skeleton wiring — it doesn't need a real cluster.
func TestEngineLifecycle(t *testing.T) {
	cp := cluster.NewExistingClusterProvider("/dev/null")
	am := agent.NewPodManager(cp.Kubeconfig())
	dc := diagnostics.NewCollector(cp.Kubeconfig())
	eng := scenario.NewEngine(cp, am, dc)

	if eng == nil {
		t.Fatal("engine should not be nil")
	}
}

// TestClusterProviderInterface verifies both provider implementations
// satisfy the Provider interface at compile time.
func TestClusterProviderInterface(t *testing.T) {
	var _ cluster.Provider = (*cluster.KindProvider)(nil)
	var _ cluster.Provider = (*cluster.ExistingClusterProvider)(nil)
}

// TestAgentManagerInterface verifies PodManager satisfies the Manager interface.
func TestAgentManagerInterface(t *testing.T) {
	var _ agent.Manager = (*agent.PodManager)(nil)
}

// TestEngineRunScenario runs the engine against a parsed scenario.
// With stub implementations this mainly tests that the wiring doesn't panic.
func TestEngineRunScenario(t *testing.T) {
	if os.Getenv("KUBE_AGENTS_INTEGRATION") == "" {
		t.Skip("skipping integration test (set KUBE_AGENTS_INTEGRATION=1 to run)")
	}

	cp := cluster.NewExistingClusterProvider(os.Getenv("KUBECONFIG"))
	am := agent.NewPodManager(cp.Kubeconfig())
	dc := diagnostics.NewCollector(cp.Kubeconfig())
	eng := scenario.NewEngine(cp, am, dc)

	scenarioDir := filepath.Join("scenarios")
	scenarios, err := scenario.LoadDir(scenarioDir)
	if err != nil {
		t.Fatalf("failed to load scenarios: %v", err)
	}

	ctx := context.Background()
	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			result := eng.Run(ctx, s)
			if !result.Passed {
				t.Errorf("scenario %q failed in %s: %v", s.Name, result.Duration, result.Error)
				if result.Diagnostics != nil {
					for agent, logs := range result.Diagnostics.AgentLogs {
						t.Logf("agent %s logs:\n%s", agent, logs)
					}
				}
			}
		})
	}
}
