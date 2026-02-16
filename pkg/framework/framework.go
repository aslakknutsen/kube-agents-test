package framework

import (
	"context"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/assertion"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Framework ties together all components and provides helpers for go test.
type Framework struct {
	ClusterProvider cluster.Provider
	AgentManager    agent.Manager
	ScenarioEngine  *Engine
	Kubeconfig      string
}

// Config holds options for setting up the framework.
type Config struct {
	ClusterProvider cluster.Provider
	PollInterval    time.Duration // defaults to 2s if zero
}

// New creates a Framework, provisions the cluster, and wires up components.
func New(ctx context.Context, cfg Config) (*Framework, error) {
	kubeconfig, err := cfg.ClusterProvider.Create(ctx)
	if err != nil {
		return nil, err
	}

	pollInterval := cfg.PollInterval
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	agentMgr := agent.NewDefaultManager(kubeconfig)
	asserter := assertion.NewPollingAsserter(kubeconfig, pollInterval)
	diag := diagnostics.NewCollector(kubeconfig)
	engine := NewEngine(kubeconfig, agentMgr, asserter, diag)

	return &Framework{
		ClusterProvider: cfg.ClusterProvider,
		AgentManager:    agentMgr,
		ScenarioEngine:  engine,
		Kubeconfig:      kubeconfig,
	}, nil
}

// RunScenario loads a scenario YAML file and runs it, reporting results via t.
func (f *Framework) RunScenario(t *testing.T, scenarioPath string) {
	t.Helper()

	s, err := scenario.ParseFile(scenarioPath)
	if err != nil {
		t.Fatalf("failed to parse scenario %s: %v", scenarioPath, err)
	}

	result := f.ScenarioEngine.Run(context.Background(), s)

	if !result.Passed {
		t.Errorf("scenario %q failed after %s: %v", result.ScenarioName, result.Duration, result.Error)
		if result.Diagnostics != nil {
			for name, logs := range result.Diagnostics.AgentLogs {
				t.Logf("agent %s logs:\n%s", name, logs)
			}
			if result.Diagnostics.Events != "" {
				t.Logf("cluster events:\n%s", result.Diagnostics.Events)
			}
		}
	}
}

// Teardown destroys the cluster and cleans up resources.
func (f *Framework) Teardown(ctx context.Context) error {
	if err := f.AgentManager.StopAll(ctx); err != nil {
		return err
	}
	return f.ClusterProvider.Destroy(ctx)
}
