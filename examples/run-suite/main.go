// Run multiple scenarios from a directory with FailFast and diagnostics on failure.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kube-agents/kube-agents-test/internal/stub"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
	"k8s.io/client-go/rest"
)

type suiteCluster struct{}

func (suiteCluster) ID() string                        { return "suite" }
func (suiteCluster) RESTConfig() (*rest.Config, error) { return &rest.Config{}, nil }
func (suiteCluster) KubeconfigPath() (string, bool)    { return "", false }

type suiteProvider struct{}

func (suiteProvider) Ensure(ctx context.Context, cfg cluster.Config) (cluster.Cluster, error) {
	return suiteCluster{}, nil
}
func (suiteProvider) Teardown(ctx context.Context, c cluster.Cluster) error { return nil }

type suiteManager struct{}

func (suiteManager) Deploy(ctx context.Context, c cluster.Cluster, agents scenario.AgentSet) error {
	return nil
}
func (suiteManager) Start(ctx context.Context, agentID string) error { return nil }
func (suiteManager) Stop(ctx context.Context, agentID string) error  { return nil }
func (suiteManager) Kill(ctx context.Context, agentID string) error  { return nil }
func (suiteManager) Teardown(ctx context.Context) error              { return nil }
func (suiteManager) SetResourceLimits(ctx context.Context, agentID string, limits agent.ResourceLimits) error {
	return nil
}
func (suiteManager) ClearResourceLimits(ctx context.Context, agentID string) error { return nil }
func (suiteManager) ApplyNetworkPolicy(ctx context.Context, spec agent.NetworkPolicySpec) (string, error) {
	return "", nil
}
func (suiteManager) RemoveNetworkPolicy(ctx context.Context, policyID string) error { return nil }

type suiteEngine struct {
	failNames map[string]bool
}

func (e suiteEngine) Run(ctx context.Context, in engine.RunInput) (engine.Result, error) {
	passed := !e.failNames[in.Scenario.Name]
	result := engine.Result{ScenarioName: in.Scenario.Name, Passed: passed}
	if !passed {
		result.Failure = &engine.FailureContext{Scenario: in.Scenario}
	}
	return result, nil
}

func main() {
	eng := suiteEngine{failNames: map[string]bool{"agent-restart-recovery": true}}

	r := runner.NewDefault(runner.Dependencies{
		Provider:  suiteProvider{},
		Manager:   suiteManager{},
		Engine:    eng,
		Collector: stub.NewCollector(),
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{
			"examples/scenarios/scaling-quota/scenario.yaml",
			"examples/scenarios/reconcile-only/scenario.yaml",
			"examples/scenarios/agent-restart/scenario.yaml",
			"examples/scenarios/fault-kill-agent/scenario.yaml",
		},
		FailFast:         true,
		SandboxNamespace: "test-agents",
		AllowFaults:      true,
		ArtifactsDir:     os.TempDir(),
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{
				Backend:     "kind",
				ClusterName: "suite",
			},
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "infrastructure error: %v\n", err)
		os.Exit(report.ExitCode())
	}

	for _, result := range report.Results {
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		fmt.Printf("[%s] %s\n", status, result.ScenarioName)
	}
	fmt.Printf("ran %d scenario(s), %d failed, exit code %d\n",
		len(report.Results), report.FailedCount(), report.ExitCode())

	if len(report.Diagnostics) > 0 {
		fmt.Println("diagnostics collected for:")
		for name := range report.Diagnostics {
			fmt.Printf("  - %s\n", name)
		}
	}

	os.Exit(report.ExitCode())
}
