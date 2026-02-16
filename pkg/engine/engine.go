// Package engine implements the scenario execution logic.
package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Result holds the outcome of running a single scenario.
type Result struct {
	Scenario    string
	Passed      bool
	Duration    time.Duration
	Error       error
	Diagnostics *diagnostics.Report
}

// Engine executes test scenarios against a Kubernetes cluster.
type Engine struct {
	KubeconfigPath string
	AgentManager   agent.Manager
	Collector      diagnostics.Collector
}

// Run executes a single scenario and returns the result.
func (e *Engine) Run(ctx context.Context, s *scenario.Scenario) *Result {
	start := time.Now()
	result := &Result{Scenario: s.Name}

	timeout := s.Timeout.Duration
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 1. Deploy agents.
	for _, name := range s.Agents {
		if err := e.AgentManager.Deploy(ctx, name); err != nil {
			result.Error = fmt.Errorf("deploying agent %s: %w", name, err)
			result.Duration = time.Since(start)
			return result
		}
	}
	defer e.AgentManager.StopAll(ctx)

	// 2. Apply initial state.
	if err := e.applySetup(ctx, &s.Setup); err != nil {
		result.Error = fmt.Errorf("applying setup: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// 3. Fire trigger.
	if s.Trigger != nil {
		if err := e.applyTrigger(ctx, s.Trigger); err != nil {
			result.Error = fmt.Errorf("applying trigger: %w", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// 4. Poll for expected state.
	if err := e.waitForExpectations(ctx, s.Expect); err != nil {
		result.Error = err
		result.Diagnostics = e.collectDiagnostics(ctx, s)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

func (e *Engine) applySetup(ctx context.Context, setup *scenario.Setup) error {
	for _, manifest := range setup.Manifests {
		if err := e.applyManifest(ctx, manifest); err != nil {
			return fmt.Errorf("applying manifest %s: %w", manifest, err)
		}
	}
	return nil
}

func (e *Engine) applyManifest(ctx context.Context, path string) error {
	// TODO: Read YAML file and use the dynamic client to create resources.
	return nil
}

func (e *Engine) applyTrigger(ctx context.Context, trigger *scenario.Trigger) error {
	if trigger.Patch != nil {
		return e.applyPatch(ctx, trigger.Patch)
	}
	return nil
}

func (e *Engine) applyPatch(ctx context.Context, patch *scenario.ResourcePatch) error {
	// TODO: Use the dynamic client to patch the specified resource.
	return nil
}

func (e *Engine) waitForExpectations(ctx context.Context, expectations []scenario.Expectation) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for expectations: %w", ctx.Err())
		case <-ticker.C:
			allMet := true
			for _, exp := range expectations {
				if !e.checkExpectation(ctx, &exp) {
					allMet = false
					break
				}
			}
			if allMet {
				return nil
			}
		}
	}
}

func (e *Engine) checkExpectation(ctx context.Context, exp *scenario.Expectation) bool {
	// TODO: Use the dynamic client to fetch the resource and check conditions.
	return false
}

func (e *Engine) collectDiagnostics(ctx context.Context, s *scenario.Scenario) *diagnostics.Report {
	if e.Collector == nil {
		return nil
	}

	report, err := e.Collector.Collect(ctx, diagnostics.Scope{
		Namespace:  e.inferNamespace(s),
		AgentNames: s.Agents,
	})
	if err != nil {
		return &diagnostics.Report{
			CollectionError: err.Error(),
		}
	}
	return report
}

func (e *Engine) inferNamespace(s *scenario.Scenario) string {
	// Use the namespace from the first expectation's resource if available.
	for _, exp := range s.Expect {
		if exp.Resource.Namespace != "" {
			return exp.Resource.Namespace
		}
	}
	return "default"
}
