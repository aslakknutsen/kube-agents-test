package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
)

func writeScenario(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func minimalScenarioYAML(name, agents string) string {
	return `name: ` + name + `
agents:
  - ` + agents + `
setup:
  manifests:
    - f.yaml
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: p
        namespace: sandbox
      conditions:
        - path: .metadata.name
          value: p
  timeout: 30s
`
}

func TestRunnerRejectsIncompleteDependencies(t *testing.T) {
	r := runner.NewDefault(runner.Dependencies{})
	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", minimalScenarioYAML("one", "a"))

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("Run() error = %v, want incomplete dependencies", err)
	}
}

func TestRunnerValidateWithRejectsOutOfSandboxBeforeEnsure(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	dir := t.TempDir()
	writeScenario(t, dir, "bad-ns.yaml", strings.Replace(
		minimalScenarioYAML("bad-ns", "a"),
		"namespace: sandbox",
		"namespace: kube-system",
		1,
	))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths:            []string{dir},
		SandboxNamespace: "sandbox",
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
		AgentConfig: agent.Config{Namespace: "sandbox"},
	})
	if err == nil {
		t.Fatal("Run() should reject scenario outside sandbox before Ensure")
	}
	if provider.ensureCalls != 0 {
		t.Errorf("Ensure called %d times, want 0 when validation fails early", provider.ensureCalls)
	}
}

func TestRunnerRedeploysWhenAgentSetChanges(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	dir := t.TempDir()
	writeScenario(t, dir, "aaa-first.yaml", minimalScenarioYAML("aaa-first", "agent-a"))
	writeScenario(t, dir, "bbb-second.yaml", minimalScenarioYAML("bbb-second", "agent-b"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
		AgentConfig: agent.Config{Namespace: "sandbox"},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if manager.deployCalls != 2 {
		t.Errorf("Deploy calls = %d, want 2 when agent sets differ", manager.deployCalls)
	}
	if manager.teardownCalls != 1 {
		t.Errorf("Teardown calls = %d, want 1 between agent set changes", manager.teardownCalls)
	}
	if eng.runCalls != 2 {
		t.Errorf("Engine.Run calls = %d, want 2", eng.runCalls)
	}
}

func TestRunnerSkipsRedeployWhenAgentSetUnchanged(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	dir := t.TempDir()
	writeScenario(t, dir, "aaa-first.yaml", minimalScenarioYAML("aaa-first", "agent-a"))
	writeScenario(t, dir, "bbb-second.yaml", minimalScenarioYAML("bbb-second", "agent-a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if manager.deployCalls != 1 {
		t.Errorf("Deploy calls = %d, want 1 when agent set unchanged", manager.deployCalls)
	}
}

func TestRunnerFailFastStopsAfterFirstFailure(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{failFirst: true}

	dir := t.TempDir()
	writeScenario(t, dir, "aaa-first.yaml", minimalScenarioYAML("aaa-first", "a"))
	writeScenario(t, dir, "bbb-second.yaml", minimalScenarioYAML("bbb-second", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths:    []string{dir},
		FailFast: true,
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if eng.runCalls != 1 {
		t.Errorf("Engine.Run calls = %d, want 1 with FailFast", eng.runCalls)
	}
	if len(report.Results) != 1 {
		t.Errorf("Results len = %d, want 1", len(report.Results))
	}
}

func TestRunnerLeaveClusterSkipsTeardown(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", minimalScenarioYAML("one", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths:        []string{dir},
		LeaveCluster: true,
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeAttached,
			Attached: &cluster.AttachedConfig{
				KubeconfigPath: "/tmp/fake",
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if provider.teardownCalls != 0 {
		t.Errorf("Teardown calls = %d, want 0 when LeaveCluster=true", provider.teardownCalls)
	}
}

func TestRunnerAttachedLeaveRunningSkipsTeardown(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", minimalScenarioYAML("one", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeAttached,
			Attached: &cluster.AttachedConfig{
				KubeconfigPath: "/tmp/fake",
				LeaveRunning:   true,
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if provider.teardownCalls != 0 {
		t.Errorf("Teardown calls = %d, want 0 when Attached.LeaveRunning=true", provider.teardownCalls)
	}
}

func TestRunnerCollectsDiagnosticsOnFailure(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{failFirst: true}
	collector := &recordingCollector{}

	dir := t.TempDir()
	writeScenario(t, dir, "fail.yaml", minimalScenarioYAML("fail-scenario", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider:  provider,
		Manager:   manager,
		Engine:    eng,
		Collector: collector,
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if collector.calls != 1 {
		t.Errorf("Collector.Collect calls = %d, want 1", collector.calls)
	}
	artifacts, ok := report.Diagnostics["fail-scenario"]
	if !ok {
		t.Fatal("expected diagnostics entry for failed scenario")
	}
	if artifacts.ScenarioName != "fail-scenario" {
		t.Errorf("artifacts.ScenarioName = %q", artifacts.ScenarioName)
	}
}

func TestReportPassedFalseWhenEmpty(t *testing.T) {
	var report runner.Report
	if report.Passed() {
		t.Error("Passed() should be false for empty report")
	}
}

func TestReportExitCodeDoesNotDistinguishInfraFailure(t *testing.T) {
	report := runner.Report{
		Results: []engine.Result{{Passed: false}},
	}
	if code := report.ExitCode(); code != 1 {
		t.Errorf("ExitCode() = %d, want 1 for failed scenario", code)
	}
}

type recordingCollector struct {
	calls int
}

func (c *recordingCollector) Collect(ctx context.Context, fc engine.FailureContext) (diagnostics.Artifacts, error) {
	c.calls++
	return diagnostics.Artifacts{ScenarioName: fc.Scenario.Name}, nil
}

func TestRunnerAgentSetOrderMatters(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	dir := t.TempDir()
	writeScenario(t, dir, "aaa.yaml", minimalScenarioYAML("aaa", "agent-a"))
	yamlB := minimalScenarioYAML("bbb", "agent-a")
	yamlB = strings.Replace(yamlB, "- agent-a", "- agent-b\n  - agent-a", 1)
	writeScenario(t, dir, "bbb.yaml", yamlB)

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if manager.deployCalls != 2 {
		t.Errorf("Deploy calls = %d, want 2 when agent order changes", manager.deployCalls)
	}
}

func TestRunnerRejectsFaultWithoutAllowFaults(t *testing.T) {
	provider := &recordingProvider{}
	dir := t.TempDir()
	content := minimalScenarioYAML("faulty", "a") + `
trigger:
  fault:
    type: kill-agent
`
	writeScenario(t, dir, "fault.yaml", content)

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  &recordingManager{},
		Engine:   &recordingEngine{},
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths:            []string{dir},
		SandboxNamespace: "sandbox",
		AllowFaults:      false,
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err == nil {
		t.Fatal("Run() should reject fault trigger when AllowFaults=false")
	}
	if provider.ensureCalls != 0 {
		t.Error("Ensure should not run when validation fails")
	}
}

func TestRunnerEngineInfraErrorReturnsPartialReport(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{infraError: true}

	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", minimalScenarioYAML("one", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err == nil {
		t.Fatal("expected infrastructure error from engine")
	}
	if len(report.Results) != 1 {
		t.Errorf("partial Results len = %d, want 1", len(report.Results))
	}
}
