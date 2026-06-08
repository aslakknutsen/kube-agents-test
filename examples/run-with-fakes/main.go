// Run a scenario with inline fake backends — exercises orchestration without a cluster.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
	"k8s.io/client-go/rest"
)

type fakeCluster struct{}

func (fakeCluster) ID() string                        { return "fake" }
func (fakeCluster) RESTConfig() (*rest.Config, error) { return &rest.Config{}, nil }
func (fakeCluster) KubeconfigPath() (string, bool)    { return "", false }

type fakeProvider struct{}

func (fakeProvider) Ensure(ctx context.Context, cfg cluster.Config) (cluster.Cluster, error) {
	return fakeCluster{}, nil
}
func (fakeProvider) Teardown(ctx context.Context, c cluster.Cluster) error { return nil }

type fakeManager struct{}

func (fakeManager) Deploy(ctx context.Context, c cluster.Cluster, agents scenario.AgentSet) error {
	fmt.Printf("deploy agents: %v\n", agents)
	return nil
}
func (fakeManager) Start(ctx context.Context, agentID string) error   { return nil }
func (fakeManager) Stop(ctx context.Context, agentID string) error    { return nil }
func (fakeManager) Kill(ctx context.Context, agentID string) error    { return nil }
func (fakeManager) Teardown(ctx context.Context) error                { fmt.Println("teardown agents"); return nil }
func (fakeManager) SetResourceLimits(ctx context.Context, agentID string, limits agent.ResourceLimits) error {
	return nil
}
func (fakeManager) ClearResourceLimits(ctx context.Context, agentID string) error { return nil }
func (fakeManager) ApplyNetworkPolicy(ctx context.Context, spec agent.NetworkPolicySpec) (string, error) {
	return "", nil
}
func (fakeManager) RemoveNetworkPolicy(ctx context.Context, policyID string) error { return nil }

type fakeEngine struct{}

func (fakeEngine) Run(ctx context.Context, in engine.RunInput) (engine.Result, error) {
	fmt.Printf("engine run: %s (%d assertions, timeout %s)\n",
		in.Scenario.Name, len(in.Scenario.Expect.Assertions), in.Scenario.Expect.Timeout)
	return engine.Result{ScenarioName: in.Scenario.Name, Passed: true}, nil
}

func main() {
	r := runner.NewDefault(runner.Dependencies{
		Provider: fakeProvider{},
		Manager:  fakeManager{},
		Engine:   fakeEngine{},
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{"examples/scenarios/scaling-quota/scenario.yaml"},
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{
				Backend:     "kind",
				ClusterName: "example",
			},
		},
		SandboxNamespace: "test-agents",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		os.Exit(report.ExitCode())
	}

	fmt.Printf("scenarios: %d passed, exit code %d\n", len(report.Results), report.ExitCode())
	if !report.Passed() {
		os.Exit(1)
	}
}
