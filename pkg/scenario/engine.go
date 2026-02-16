package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/assertion"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
)

// Result holds the outcome of running a single scenario.
type Result struct {
	Scenario    string
	Passed      bool
	Duration    time.Duration
	Error       error
	Diagnostics *diagnostics.Report
}

// Engine executes test scenarios against a cluster.
type Engine struct {
	clusterProvider cluster.Provider
	agentManager    agent.Manager
	collector       *diagnostics.Collector
}

func NewEngine(cp cluster.Provider, am agent.Manager, dc *diagnostics.Collector) *Engine {
	return &Engine{
		clusterProvider: cp,
		agentManager:    am,
		collector:       dc,
	}
}

// Run executes a single scenario and returns the result.
func (e *Engine) Run(ctx context.Context, s *Scenario) *Result {
	start := time.Now()
	result := &Result{Scenario: s.Name}

	if err := e.setup(ctx, s); err != nil {
		result.Error = fmt.Errorf("setup: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	if err := e.deployAgents(ctx, s); err != nil {
		result.Error = fmt.Errorf("deploying agents: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	if err := e.fireTrigger(ctx, s); err != nil {
		result.Error = fmt.Errorf("trigger: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	timeout := s.Expect.Timeout.Duration
	if timeout == 0 {
		timeout = 2 * time.Minute
	}

	assertCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := e.waitForExpectedState(assertCtx, s); err != nil {
		result.Error = fmt.Errorf("assertion: %w", err)
		result.Diagnostics = e.collectDiagnostics(ctx, s)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

func (e *Engine) setup(ctx context.Context, s *Scenario) error {
	// TODO: apply each manifest in s.Setup.Manifests using client-go
	for _, m := range s.Setup.Manifests {
		_ = m // apply manifest
	}
	return nil
}

func (e *Engine) deployAgents(ctx context.Context, s *Scenario) error {
	for _, name := range s.Agents {
		if err := e.agentManager.Deploy(ctx, name); err != nil {
			return fmt.Errorf("deploying agent %s: %w", name, err)
		}
	}
	return nil
}

func (e *Engine) fireTrigger(ctx context.Context, s *Scenario) error {
	if s.Trigger == nil {
		return nil
	}
	// TODO: apply the patch trigger using client-go dynamic client
	return nil
}

func (e *Engine) waitForExpectedState(ctx context.Context, s *Scenario) error {
	for _, exp := range s.Expect.Resources {
		checker := assertion.ResourceChecker{
			APIVersion: exp.APIVersion,
			Kind:       exp.Kind,
			Name:       exp.Name,
			Namespace:  exp.Namespace,
			Conditions: toAssertionConditions(exp.Conditions),
		}
		if err := assertion.PollUntilMatch(ctx, checker); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) collectDiagnostics(ctx context.Context, s *Scenario) *diagnostics.Report {
	if e.collector == nil {
		return nil
	}
	report, _ := e.collector.Collect(ctx, s.Name, s.Agents)
	return report
}

func toAssertionConditions(conds []Condition) []assertion.Condition {
	out := make([]assertion.Condition, len(conds))
	for i, c := range conds {
		out[i] = assertion.Condition{Path: c.Path, Value: c.Value}
	}
	return out
}
