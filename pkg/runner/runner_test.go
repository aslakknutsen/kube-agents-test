package runner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

type fakeCluster struct {
	provisioned bool
}

func (f *fakeCluster) Provision(ctx context.Context) (*cluster.Cluster, error) {
	_ = ctx
	f.provisioned = true
	return &cluster.Cluster{KubeconfigPath: "/tmp/fake"}, nil
}

func (f *fakeCluster) Attach(ctx context.Context, kubeconfig string) (*cluster.Cluster, error) {
	_ = ctx
	return &cluster.Cluster{KubeconfigPath: kubeconfig}, nil
}

func (f *fakeCluster) Teardown(ctx context.Context) error {
	_ = ctx
	return nil
}

type fakeAgent struct {
	deployed  []string
	deployErr error
}

func (f *fakeAgent) DeploySet(ctx context.Context, agents []string) error {
	_ = ctx
	if f.deployErr != nil {
		return f.deployErr
	}
	f.deployed = append([]string(nil), agents...)
	return nil
}

func (f *fakeAgent) Start(ctx context.Context, name string) error   { _ = ctx; _ = name; return nil }
func (f *fakeAgent) Stop(ctx context.Context, name string) error    { _ = ctx; _ = name; return nil }
func (f *fakeAgent) Kill(ctx context.Context, name string) error    { _ = ctx; _ = name; return nil }
func (f *fakeAgent) ApplyDegraded(ctx context.Context, name string, opts agent.DegradedOptions) error {
	_ = ctx
	_ = name
	_ = opts
	return nil
}
func (f *fakeAgent) Teardown(ctx context.Context) error { _ = ctx; return nil }

type fakeEngine struct {
	pass bool
}

func (f *fakeEngine) Run(ctx context.Context, sc *scenario.Scenario, opts scenario.RunOptions) (*scenario.Result, error) {
	_ = ctx
	_ = opts
	res := &scenario.Result{
		ScenarioName: sc.Name,
		Passed:       f.pass,
		Duration:     time.Second,
	}
	if !f.pass {
		res.Failure = &scenario.AssertionFailure{Message: "mismatch", TimedOut: true}
	}
	return res, nil
}

type fakeDiag struct{}

func (f *fakeDiag) Collect(ctx context.Context, req diagnostics.CollectRequest) (*diagnostics.Bundle, error) {
	_ = ctx
	_ = req
	return &diagnostics.Bundle{AgentLogs: map[string][]byte{"a": []byte("log")}}, nil
}

func TestRunner_runOneScenario_pass(t *testing.T) {
	cp := &fakeCluster{}
	ag := &fakeAgent{}
	eng := &fakeEngine{pass: true}

	r, err := runner.New(runner.Config{
		ClusterProvider: cp,
		AgentManager:    ag,
		ScenarioEngine:  eng,
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}

	sc := &scenario.Scenario{
		Name:   "ok",
		Agents: []string{"a"},
		Expect: scenario.Expectation{
			Resources: []scenario.ResourceExpect{{
				Resource: scenario.ObjectRef{APIVersion: "v1", Kind: "Pod", Name: "p"},
				Conditions: []scenario.Condition{{Path: ".metadata.name", Value: "p"}},
			}},
			Timeout: time.Minute,
		},
	}
	suite, err := r.Run(context.Background(), []runner.LoadedScenario{{Scenario: sc}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !cp.provisioned {
		t.Error("expected cluster provision")
	}
	if len(ag.deployed) != 1 || ag.deployed[0] != "a" {
		t.Errorf("deployed = %v", ag.deployed)
	}
	if suite.Passed != 1 || suite.Failed != 0 {
		t.Fatalf("suite = %+v", suite)
	}
}

func TestRunner_collectsDiagnosticsOnFailure(t *testing.T) {
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		AgentManager:    &fakeAgent{},
		ScenarioEngine:  &fakeEngine{pass: false},
		Diagnostics:     &fakeDiag{},
		ClusterMode:     cluster.ModeEphemeral,
	})
	if err != nil {
		t.Fatal(err)
	}

	sc := &scenario.Scenario{
		Name:   "fail",
		Agents: []string{"a"},
		Expect: scenario.Expectation{
			Resources: []scenario.ResourceExpect{{
				Resource: scenario.ObjectRef{APIVersion: "v1", Kind: "Pod", Name: "p"},
				Conditions: []scenario.Condition{{Path: ".metadata.name", Value: "p"}},
			}},
			Timeout: time.Minute,
		},
	}
	suite, err := r.Run(context.Background(), []runner.LoadedScenario{{Scenario: sc}})
	if err != nil {
		t.Fatal(err)
	}
	if suite.Failed != 1 {
		t.Fatalf("suite = %+v", suite)
	}
	res := suite.Results[0]
	if res.Diagnostics == nil {
		t.Fatal("expected diagnostics bundle")
	}
	if _, ok := res.Diagnostics.AgentLogs["a"]; !ok {
		t.Error("expected agent log key")
	}
}

func TestRunner_attachedRequiresKubeconfig(t *testing.T) {
	r, err := runner.New(runner.Config{
		ClusterProvider: &fakeCluster{},
		ClusterMode:     cluster.ModeAttached,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Run(context.Background(), []runner.LoadedScenario{{
		Scenario: &scenario.Scenario{
			Name: "x", Agents: []string{"a"},
			Expect: scenario.Expectation{
				Resources: []scenario.ResourceExpect{{
					Resource:   scenario.ObjectRef{APIVersion: "v1", Kind: "Pod", Name: "p"},
					Conditions: []scenario.Condition{{Path: ".x", Value: 1}},
				}},
				Timeout: time.Minute,
			},
		},
	}})
	if err == nil || !errors.Is(err, errors.New("runner: kubeconfig required for attached mode")) {
		// errors.Is won't work on fmt.Errorf wrapped; check substring
		if err == nil || err.Error() != "runner: kubeconfig required for attached mode" {
			t.Fatalf("err = %v", err)
		}
	}
}
