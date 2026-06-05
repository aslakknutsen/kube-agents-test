package runner_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

type sandboxRecordingEngine struct {
	recordingEngine
	lastSandbox string
}

func (e *sandboxRecordingEngine) Run(ctx context.Context, in engine.RunInput) (engine.Result, error) {
	e.runCalls++
	e.lastSandbox = in.SandboxNamespace
	return e.recordingEngine.Run(ctx, in)
}

type failingDeployManager struct {
	recordingManager
}

func (m *failingDeployManager) Deploy(ctx context.Context, c cluster.Cluster, agents scenario.AgentSet) error {
	m.deployCalls++
	return errors.New("deploy failed")
}

type failingMidTeardownManager struct {
	recordingManager
	failTeardown bool
}

func (m *failingMidTeardownManager) Teardown(ctx context.Context) error {
	m.teardownCalls++
	if m.failTeardown && m.teardownCalls == 1 {
		return errors.New("mid teardown failed")
	}
	return nil
}

func TestRunnerFailFastStillTearsDownAgents(t *testing.T) {
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

	_, err := r.Run(context.Background(), runner.RunOptions{
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
	if manager.teardownCalls != 1 {
		t.Errorf("Teardown calls = %d, want 1 after FailFast (defer final cleanup)", manager.teardownCalls)
	}
}

func TestRunnerDeployFailureSetsInfraFailed(t *testing.T) {
	provider := &recordingProvider{}
	manager := &failingDeployManager{}
	eng := &recordingEngine{}

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
		t.Fatal("expected deploy infrastructure error")
	}
	if !report.InfraFailed {
		t.Error("InfraFailed should be true when Deploy fails")
	}
	if report.ExitCode() != 2 {
		t.Errorf("ExitCode() = %d, want 2", report.ExitCode())
	}
	if eng.runCalls != 0 {
		t.Errorf("Engine.Run calls = %d, want 0 when deploy fails", eng.runCalls)
	}
	if provider.teardownCalls != 1 {
		t.Errorf("Provider Teardown calls = %d, want 1 via defer", provider.teardownCalls)
	}
}

func TestRunnerForwardsSandboxNamespaceToEngine(t *testing.T) {
	eng := &sandboxRecordingEngine{}
	dir := t.TempDir()
	yaml := strings.Replace(minimalScenarioYAML("one", "a"), "namespace: sandbox", "namespace: tenant-a", 1)
	writeScenario(t, dir, "one.yaml", yaml)

	r := runner.NewDefault(runner.Dependencies{
		Provider: &recordingProvider{},
		Manager:  &recordingManager{},
		Engine:   eng,
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
	if eng.lastSandbox != "tenant-a" {
		t.Errorf("SandboxNamespace forwarded = %q, want tenant-a", eng.lastSandbox)
	}
}

func TestRunnerEnsureFailureSkipsClusterTeardown(t *testing.T) {
	provider := &recordingProvider{ensureErr: errors.New("ensure failed")}
	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", minimalScenarioYAML("one", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  &recordingManager{},
		Engine:   &recordingEngine{},
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err == nil {
		t.Fatal("expected ensure error")
	}
	if !report.InfraFailed {
		t.Error("InfraFailed should be true when Ensure fails")
	}
	if provider.teardownCalls != 0 {
		t.Errorf("Provider Teardown calls = %d, want 0 when Ensure never succeeded", provider.teardownCalls)
	}
}

func TestRunnerMidRunTeardownFailureSetsInfraFailed(t *testing.T) {
	manager := &failingMidTeardownManager{failTeardown: true}
	dir := t.TempDir()
	writeScenario(t, dir, "aaa-first.yaml", minimalScenarioYAML("aaa-first", "agent-a"))
	writeScenario(t, dir, "bbb-second.yaml", minimalScenarioYAML("bbb-second", "agent-b"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: &recordingProvider{},
		Manager:  manager,
		Engine:   &recordingEngine{},
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err == nil {
		t.Fatal("expected mid-run teardown error")
	}
	if !strings.Contains(err.Error(), "mid teardown failed") {
		t.Fatalf("error = %v", err)
	}
	if !report.InfraFailed {
		t.Error("InfraFailed should be true on mid-run teardown failure")
	}
	if manager.teardownCalls != 2 {
		t.Errorf("Teardown calls = %d, want 2 (failed mid-run + defer final)", manager.teardownCalls)
	}
}

func TestReportExitCodeInfraPreemptsScenarioFailure(t *testing.T) {
	report := runner.Report{
		InfraFailed: true,
		Results:     []engine.Result{{Passed: false}},
	}
	if code := report.ExitCode(); code != 2 {
		t.Errorf("ExitCode() = %d, want 2 (infra preempts scenario failure)", code)
	}
}

func TestRunnerDuplicateScenarioNamesOverwriteDiagnostics(t *testing.T) {
	eng := &recordingEngine{failFirst: true}
	collector := &recordingCollector{}

	dir := t.TempDir()
	writeScenario(t, dir, "first.yaml", minimalScenarioYAML("same-name", "a"))
	writeScenario(t, dir, "second.yaml", minimalScenarioYAML("same-name", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider:  &recordingProvider{},
		Manager:   &recordingManager{},
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
	if collector.calls != 2 {
		t.Errorf("Collector.Collect calls = %d, want 2 for two failing scenarios", collector.calls)
	}
	// Diagnostics map is keyed by scenario name, not file path — duplicate names collide.
	if len(report.Diagnostics) != 1 {
		t.Errorf("Diagnostics entries = %d, current map keys on scenario name only", len(report.Diagnostics))
	}
}

func TestRunnerOptsDepsPartialOverrideReplacesAllDeps(t *testing.T) {
	defaultEng := &recordingEngine{}
	customEng := &recordingEngine{}
	dir := t.TempDir()
	writeScenario(t, dir, "one.yaml", minimalScenarioYAML("one", "a"))

	r := runner.NewDefault(runner.Dependencies{
		Provider: &recordingProvider{},
		Manager:  &recordingManager{},
		Engine:   defaultEng,
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{dir},
		Deps: runner.Dependencies{
			Engine: customEng,
		},
		ClusterConfig: cluster.Config{
			Mode:      cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if customEng.runCalls != 0 {
		t.Errorf("custom engine Run calls = %d, want 0 when Provider nil triggers full deps replacement", customEng.runCalls)
	}
	if defaultEng.runCalls != 1 {
		t.Errorf("default engine Run calls = %d, want 1", defaultEng.runCalls)
	}
}
