package runner_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

type countingCluster struct {
	fakeCluster
	teardowns int
}

func (c *countingCluster) Teardown(ctx context.Context) error {
	_ = ctx
	c.teardowns++
	return nil
}

type countingAgent struct {
	fakeAgent
	teardowns int
}

func (a *countingAgent) Teardown(ctx context.Context) error {
	_ = ctx
	a.teardowns++
	return nil
}

type optsCapturingEngine struct {
	lastOpts scenario.RunOptions
	fakeEngine
}

func (e *optsCapturingEngine) Run(ctx context.Context, sc *scenario.Scenario, opts scenario.RunOptions) (*scenario.Result, error) {
	e.lastOpts = opts
	return e.fakeEngine.Run(ctx, sc, opts)
}

func minimalScenario(name string) *scenario.Scenario {
	return &scenario.Scenario{
		Name:   name,
		Agents: []string{"a"},
		Expect: scenario.Expectation{
			Resources: []scenario.ResourceExpect{{
				Resource:   scenario.ObjectRef{APIVersion: "v1", Kind: "Pod", Name: "p"},
				Conditions: []scenario.Condition{{Path: ".metadata.name", Value: "p"}},
			}},
			Timeout: time.Minute,
		},
	}
}

func TestRunner_emptyScenarios(t *testing.T) {
	r, err := runner.New(runner.Config{ClusterProvider: &fakeCluster{}})
	if err != nil {
		t.Fatal(err)
	}
	suite, err := r.Run(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if suite.Passed != 0 || suite.Failed != 0 || len(suite.Results) != 0 {
		t.Fatalf("suite = %+v", suite)
	}
}

func TestRunner_attachedSkipsClusterTeardown(t *testing.T) {
	cp := &countingCluster{}
	r, err := runner.New(runner.Config{
		ClusterProvider: cp,
		AgentManager:    &fakeAgent{},
		ScenarioEngine:  &fakeEngine{pass: true},
		ClusterMode:     cluster.ModeAttached,
		Kubeconfig:      "/tmp/kc",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Run(context.Background(), []*scenario.Scenario{minimalScenario("ok")})
	if err != nil {
		t.Fatal(err)
	}
	if cp.teardowns != 0 {
		t.Errorf("attached teardowns = %d, want 0", cp.teardowns)
	}
}

func TestRunner_ephemeralTeardownsCluster(t *testing.T) {
	cp := &countingCluster{}
	r, err := runner.New(runner.Config{
		ClusterProvider: cp,
		AgentManager:    &fakeAgent{},
		ScenarioEngine:  &fakeEngine{pass: true},
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Run(context.Background(), []*scenario.Scenario{minimalScenario("ok")})
	if err != nil {
		t.Fatal(err)
	}
	if cp.teardowns != 1 {
		t.Errorf("teardowns = %d, want 1", cp.teardowns)
	}
}

func TestRunner_deployFailureSkipsEngine(t *testing.T) {
	eng := &fakeEngine{pass: true}
	ag := &fakeAgent{}
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		AgentManager:    ag,
		ScenarioEngine:  eng,
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}
	ag.deployErr = errors.New("boom")
	suite, err := r.Run(context.Background(), []*scenario.Scenario{minimalScenario("x")})
	if err != nil {
		t.Fatal(err)
	}
	if suite.Failed != 1 {
		t.Fatalf("suite = %+v", suite)
	}
	if suite.Results[0].Result != nil {
		t.Error("expected no engine result on deploy failure")
	}
}

func TestRunner_multipleScenariosTeardownAgentsEachTime(t *testing.T) {
	ag := &countingAgent{}
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		AgentManager:    ag,
		ScenarioEngine:  &fakeEngine{pass: true},
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Run(context.Background(), []*scenario.Scenario{
		minimalScenario("one"),
		minimalScenario("two"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ag.teardowns != 2 {
		t.Errorf("agent teardowns = %d, want 2", ag.teardowns)
	}
}

func TestRunner_diagnosticsNotImplementedDoesNotFailRun(t *testing.T) {
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		AgentManager:    &fakeAgent{},
		ScenarioEngine:  &fakeEngine{pass: false},
		Diagnostics:     &notImplDiag{},
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}
	suite, err := r.Run(context.Background(), []*scenario.Scenario{minimalScenario("fail")})
	if err != nil {
		t.Fatal(err)
	}
	res := suite.Results[0]
	if res.Err != nil {
		t.Fatalf("unexpected infra err: %v", res.Err)
	}
	if res.Diagnostics != nil {
		t.Fatal("expected nil bundle when collector returns ErrNotImplemented")
	}
}

func TestRunner_engineNilResultMarksFailed(t *testing.T) {
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		AgentManager:    &fakeAgent{},
		ScenarioEngine:  &nilResultEngine{},
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}
	suite, err := r.Run(context.Background(), []*scenario.Scenario{minimalScenario("nil")})
	if err != nil {
		t.Fatal(err)
	}
	if suite.Failed != 1 || suite.Results[0].Passed {
		t.Fatalf("result = %+v", suite.Results[0])
	}
}

func TestRunner_runOptionsBaseDirNotSetYet(t *testing.T) {
	eng := &optsCapturingEngine{fakeEngine: fakeEngine{pass: true}}
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		AgentManager:    &fakeAgent{},
		ScenarioEngine:  eng,
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Run(context.Background(), []*scenario.Scenario{minimalScenario("x")})
	if err != nil {
		t.Fatal(err)
	}
	if eng.lastOpts.BaseDir != "" {
		t.Errorf("BaseDir = %q, want empty until runner tracks scenario paths", eng.lastOpts.BaseDir)
	}
}

func TestLoadPath_fileAndDirectory(t *testing.T) {
	dir := filepath.Join("..", "scenario", "testdata")
	fromDir, err := runner.LoadPath(dir)
	if err != nil {
		t.Fatalf("LoadPath dir: %v", err)
	}
	if len(fromDir) < 2 {
		t.Fatalf("dir len = %d", len(fromDir))
	}
	file := filepath.Join(dir, "example-scenario.yaml")
	fromFile, err := runner.LoadPath(file)
	if err != nil {
		t.Fatalf("LoadPath file: %v", err)
	}
	if len(fromFile) != 1 {
		t.Fatalf("file len = %d", len(fromFile))
	}
}

type notImplDiag struct{}

func (notImplDiag) Collect(ctx context.Context, req diagnostics.CollectRequest) (*diagnostics.Bundle, error) {
	_ = ctx
	_ = req
	return nil, diagnostics.ErrNotImplemented
}

type nilResultEngine struct{}

func (nilResultEngine) Run(ctx context.Context, sc *scenario.Scenario, opts scenario.RunOptions) (*scenario.Result, error) {
	_ = ctx
	_ = sc
	_ = opts
	return nil, nil
}
