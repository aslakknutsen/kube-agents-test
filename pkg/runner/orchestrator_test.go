package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kube-agents/kube-agents-test/internal/errs"
	"github.com/kube-agents/kube-agents-test/internal/stub"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
	"k8s.io/client-go/rest"
)

type fakeCluster struct{}

func (fakeCluster) ID() string                          { return "fake" }
func (fakeCluster) RESTConfig() (*rest.Config, error)   { return nil, nil }
func (fakeCluster) KubeconfigPath() (string, bool)      { return "", false }

type recordingProvider struct {
	ensureCalls   int
	teardownCalls int
}

func (p *recordingProvider) Ensure(ctx context.Context, cfg cluster.Config) (cluster.Cluster, error) {
	p.ensureCalls++
	return fakeCluster{}, nil
}

func (p *recordingProvider) Teardown(ctx context.Context, c cluster.Cluster) error {
	p.teardownCalls++
	return nil
}

type recordingManager struct {
	deployCalls   int
	teardownCalls int
}

func (m *recordingManager) Deploy(ctx context.Context, c cluster.Cluster, agents scenario.AgentSet) error {
	m.deployCalls++
	return nil
}

func (m *recordingManager) Start(ctx context.Context, agentID string) error   { return nil }
func (m *recordingManager) Stop(ctx context.Context, agentID string) error    { return nil }
func (m *recordingManager) Kill(ctx context.Context, agentID string) error    { return nil }
func (m *recordingManager) Teardown(ctx context.Context) error                  { m.teardownCalls++; return nil }
func (m *recordingManager) SetResourceLimits(ctx context.Context, agentID string, limits agent.ResourceLimits) error {
	return nil
}
func (m *recordingManager) ClearResourceLimits(ctx context.Context, agentID string) error { return nil }
func (m *recordingManager) ApplyNetworkPolicy(ctx context.Context, spec agent.NetworkPolicySpec) (string, error) {
	return "", nil
}
func (m *recordingManager) RemoveNetworkPolicy(ctx context.Context, policyID string) error { return nil }

type recordingEngine struct {
	runCalls int
}

func (e *recordingEngine) Run(ctx context.Context, in engine.RunInput) (engine.Result, error) {
	e.runCalls++
	return engine.Result{ScenarioName: in.Scenario.Name, Passed: true}, nil
}

func TestRunnerOrchestrationOrder(t *testing.T) {
	provider := &recordingProvider{}
	manager := &recordingManager{}
	eng := &recordingEngine{}

	r := runner.NewDefault(runner.Dependencies{
		Provider: provider,
		Manager:  manager,
		Engine:   eng,
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{"../../testdata/scenarios"},
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{
				Backend:     "kind",
				ClusterName: "test",
			},
		},
		AgentConfig: agent.Config{Namespace: "test"},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if provider.ensureCalls != 1 {
		t.Errorf("Ensure calls = %d, want 1", provider.ensureCalls)
	}
	if provider.teardownCalls != 1 {
		t.Errorf("Teardown calls = %d, want 1", provider.teardownCalls)
	}
	if manager.deployCalls != 1 {
		t.Errorf("Deploy calls = %d, want 1", manager.deployCalls)
	}
	if eng.runCalls != 1 {
		t.Errorf("Engine.Run calls = %d, want 1", eng.runCalls)
	}
	if !report.Passed() {
		t.Errorf("report.Passed() = false, want true")
	}
}

func TestRunnerStubProviderReturnsNotImplemented(t *testing.T) {
	r := runner.NewDefault(runner.Dependencies{
		Provider: stub.NewProvider(),
		Manager:  stub.NewManager(),
		Engine:   stub.NewEngine(),
	})

	_, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{"../../testdata/scenarios"},
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{Backend: "kind"},
		},
	})
	if !errors.Is(err, errs.ErrNotImplemented) {
		t.Fatalf("Run() error = %v, want ErrNotImplemented", err)
	}
}
