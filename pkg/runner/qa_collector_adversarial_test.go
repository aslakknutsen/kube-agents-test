package runner_test

import (
	"context"
	"strings"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
)

type sandboxCollector struct {
	lastSandbox string
}

func (c *sandboxCollector) Collect(ctx context.Context, fc engine.FailureContext) (diagnostics.Artifacts, error) {
	c.lastSandbox = fc.SandboxNamespace
	return diagnostics.Artifacts{ScenarioName: fc.Scenario.Name}, nil
}

// Documents current gap: runner forwards SandboxNamespace to Engine.RunInput but not
// to FailureContext before Collector.Collect (ArtifactsDir is forwarded).
func TestRunnerFailureContextSandboxNamespaceNotForwardedToCollector(t *testing.T) {
	collector := &sandboxCollector{}
	dir := t.TempDir()
	yaml := strings.Replace(minimalScenarioYAML("fail-scenario", "a"), "namespace: sandbox", "namespace: tenant-a", 1)
	writeScenario(t, dir, "fail.yaml", yaml)

	r := runner.NewDefault(runner.Dependencies{
		Provider:  &recordingProvider{},
		Manager:   &recordingManager{},
		Engine:    &recordingEngine{failFirst: true},
		Collector: collector,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths:            []string{dir},
		SandboxNamespace: "tenant-a",
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if collector.lastSandbox != "" {
		t.Errorf("FailureContext.SandboxNamespace = %q, runner does not populate it today", collector.lastSandbox)
	}
}
