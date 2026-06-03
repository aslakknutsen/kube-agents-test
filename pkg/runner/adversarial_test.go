package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestRunNilDependencies(t *testing.T) {
	r := runner.New(runner.Config{}, runner.Dependencies{})
	_, err := r.Run(context.Background(), []*scenario.Scenario{{Name: "x"}})
	if err == nil {
		t.Fatal("expected error for nil dependencies")
	}
}

func TestRunNilScenarioReturnsErrorStatus(t *testing.T) {
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &okCluster{},
		AgentManager:    &okAgent{},
		Engine:          &okEngine{},
	})
	report, err := r.Run(context.Background(), []*scenario.Scenario{nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("results = %d", len(report.Results))
	}
	if report.Results[0].Status != engine.StatusError {
		t.Fatalf("status = %q, want error", report.Results[0].Status)
	}
	if report.Summary.Errored != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

func TestRunStubClusterReturnsErrorStatus(t *testing.T) {
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &failCluster{err: errors.New("cluster down")},
		AgentManager:    &okAgent{},
		Engine:          &okEngine{},
	})
	sc := validScenario(t)
	report, err := r.Run(context.Background(), []*scenario.Scenario{sc})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Errored != 1 || report.Summary.Passed != 0 {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if report.Results[0].Status != engine.StatusError {
		t.Fatalf("status = %q", report.Results[0].Status)
	}
}

func TestRunSuccessfulMock(t *testing.T) {
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &okCluster{},
		AgentManager:    &okAgent{},
		Engine:          &okEngine{},
	})
	sc := validScenario(t)
	report, err := r.Run(context.Background(), []*scenario.Scenario{sc})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Passed != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

func TestRunPathEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &okCluster{},
		AgentManager:    &okAgent{},
		Engine:          &okEngine{},
	})
	_, err := r.RunPath(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error for empty directory")
	}
}

func validScenario(t *testing.T) *scenario.Scenario {
	t.Helper()
	return &scenario.Scenario{
		Name:   "test",
		Agents: []string{"a"},
		Setup:  scenario.Setup{Manifests: []string{"m.yaml"}},
		Expect: scenario.Expect{
			Timeout: 1,
			Assertions: []scenario.ResourceAssertion{{
				Resource: scenario.ResourceRef{
					APIVersion: "v1", Kind: "Pod", Name: "p", Namespace: "ns",
				},
				Conditions: []scenario.PathCondition{{Path: ".x", Value: 1}},
			}},
		},
	}
}

type okCluster struct{}

func (c *okCluster) Acquire(ctx context.Context) (cluster.Handle, error) {
	return &okHandle{}, nil
}

type okHandle struct{}

func (h *okHandle) Kubeconfig() ([]byte, error) { return []byte("kubeconfig"), nil }
func (h *okHandle) Release(ctx context.Context) error { return nil }

type failCluster struct{ err error }

func (c *failCluster) Acquire(ctx context.Context) (cluster.Handle, error) {
	return nil, c.err
}

type okAgent struct{}

func (a *okAgent) DeploySet(ctx context.Context, agents []string) error { return nil }
func (a *okAgent) Start(ctx context.Context, name string) error         { return nil }
func (a *okAgent) Stop(ctx context.Context, name string) error          { return nil }
func (a *okAgent) Kill(ctx context.Context, name string) error          { return nil }
func (a *okAgent) ApplyDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (a *okAgent) ClearDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (a *okAgent) Teardown(ctx context.Context) error { return nil }

type okEngine struct{}

func (e *okEngine) Execute(ctx context.Context, req engine.ExecuteRequest) (*engine.Result, error) {
	return &engine.Result{
		ScenarioName: req.Scenario.Name,
		Status:       engine.StatusPassed,
	}, nil
}
