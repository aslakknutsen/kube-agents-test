package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/assertion"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Result holds the outcome of running a single scenario.
type Result struct {
	ScenarioName string
	Passed       bool
	Duration     time.Duration
	Error        error
	Diagnostics  *diagnostics.Report
}

// Engine executes test scenarios against a cluster.
type Engine struct {
	Kubeconfig   string
	AgentManager agent.Manager
	Asserter     assertion.Asserter
	Diagnostics  *diagnostics.Collector
}

// NewEngine creates a scenario engine wired to the given components.
func NewEngine(kubeconfig string, agentMgr agent.Manager, asserter assertion.Asserter, diag *diagnostics.Collector) *Engine {
	return &Engine{
		Kubeconfig:   kubeconfig,
		AgentManager: agentMgr,
		Asserter:     asserter,
		Diagnostics:  diag,
	}
}

// Run executes a scenario end-to-end:
// 1. Apply initial state
// 2. Deploy agents
// 3. Fire trigger
// 4. Wait for expected state (poll until timeout)
// 5. Collect diagnostics on failure
func (e *Engine) Run(ctx context.Context, s *scenario.Scenario) *Result {
	start := time.Now()

	result := &Result{ScenarioName: s.Name}

	if err := e.applySetup(ctx, s.Setup); err != nil {
		result.Error = fmt.Errorf("setup: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	agentConfigs := e.buildAgentConfigs(s.Agents)
	if err := e.AgentManager.Deploy(ctx, agentConfigs); err != nil {
		result.Error = fmt.Errorf("deploying agents: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	if s.Trigger != nil {
		if err := e.applyTrigger(ctx, s.Trigger); err != nil {
			result.Error = fmt.Errorf("trigger: %w", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	timeout := s.Expect.Timeout.Duration
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := e.Asserter.WaitForState(pollCtx, s.Expect.Resources); err != nil {
		result.Error = fmt.Errorf("assertion failed: %w", err)
		result.Diagnostics = e.Diagnostics.Collect(ctx, s.Agents, "")
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

func (e *Engine) applySetup(ctx context.Context, setup scenario.Setup) error {
	// TODO: kubectl apply -f each manifest, or use client-go to create resources
	_ = setup
	return nil
}

func (e *Engine) applyTrigger(ctx context.Context, trigger *scenario.Trigger) error {
	// TODO: apply the patch via client-go dynamic client
	_ = trigger
	return nil
}

func (e *Engine) buildAgentConfigs(agentNames []string) []agent.AgentConfig {
	configs := make([]agent.AgentConfig, len(agentNames))
	for i, name := range agentNames {
		configs[i] = agent.AgentConfig{
			Name: name,
			Mode: agent.DeployModePod,
		}
	}
	return configs
}
