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

func TestRunAgentDeployFailureReleasesCluster(t *testing.T) {
	cluster := &trackingCluster{}
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: cluster,
		AgentManager:    &failAgent{err: errors.New("deploy failed")},
		Engine:          &okEngine{},
	})
	sc := validScenario(t)
	report, err := r.Run(context.Background(), []*scenario.Scenario{sc})
	if err != nil {
		t.Fatal(err)
	}
	if !cluster.released {
		t.Fatal("expected cluster release after deploy failure")
	}
	if report.Summary.Errored != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

func TestRunMultipleScenariosAggregatesSummary(t *testing.T) {
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &okCluster{},
		AgentManager:    &okAgent{},
		Engine:          &mixedEngine{},
	})
	scenarios := []*scenario.Scenario{validScenario(t), validScenario(t), validScenario(t)}
	report, err := r.Run(context.Background(), scenarios)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Total != 3 || report.Summary.Passed != 2 || report.Summary.Failed != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

type trackingCluster struct {
	released bool
}

func (c *trackingCluster) Acquire(ctx context.Context) (cluster.Handle, error) {
	return &trackingHandle{c: c}, nil
}

type trackingHandle struct{ c *trackingCluster }

func (h *trackingHandle) Kubeconfig() ([]byte, error) { return nil, nil }
func (h *trackingHandle) Release(ctx context.Context) error {
	h.c.released = true
	return nil
}

type failAgent struct{ err error }

func (a *failAgent) DeploySet(ctx context.Context, agents []string) error { return a.err }
func (a *failAgent) Start(ctx context.Context, name string) error         { return nil }
func (a *failAgent) Stop(ctx context.Context, name string) error          { return nil }
func (a *failAgent) Kill(ctx context.Context, name string) error          { return nil }
func (a *failAgent) ApplyDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (a *failAgent) ClearDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (a *failAgent) Teardown(ctx context.Context) error { return nil }

type mixedEngine struct{ n int }

func (e *mixedEngine) Execute(ctx context.Context, req engine.ExecuteRequest) (*engine.Result, error) {
	e.n++
	status := engine.StatusPassed
	if e.n == 2 {
		status = engine.StatusFailed
	}
	return &engine.Result{ScenarioName: req.Scenario.Name, Status: status}, nil
}
