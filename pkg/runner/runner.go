// Package runner defines the Test Runner API that orchestrates subsystem lifecycles.
// See docs/architecture.md#test-runner.
package runner

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Config holds runner-level settings across subsystems.
type Config struct {
	Cluster cluster.Config `yaml:"cluster"`
	Agent   agent.Config   `yaml:"agent"`
}

// Dependencies injects subsystem implementations for New.
type Dependencies struct {
	ClusterProvider cluster.Provider
	AgentManager    agent.Manager
	Engine          engine.Engine
}

// Summary aggregates counts across scenario results.
type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Errored int
}

// RunReport is the aggregate outcome of a test run.
type RunReport struct {
	Results []*engine.Result
	Summary Summary
}

// Runner orchestrates scenario execution across cluster, agent, and engine subsystems.
type Runner interface {
	Run(ctx context.Context, scenarios []*scenario.Scenario) (*RunReport, error)
	RunPath(ctx context.Context, path string) (*RunReport, error)
}

// New constructs a Runner with the given configuration and dependencies.
func New(cfg Config, deps Dependencies) Runner {
	return &defaultRunner{cfg: cfg, deps: deps}
}

type defaultRunner struct {
	cfg  Config
	deps Dependencies
}

func (r *defaultRunner) Run(ctx context.Context, scenarios []*scenario.Scenario) (*RunReport, error) {
	return runScenarios(ctx, r.deps, scenarios, "")
}

func (r *defaultRunner) RunPath(ctx context.Context, path string) (*RunReport, error) {
	scenarios, bases, err := loadScenariosFromPath(path)
	if err != nil {
		return nil, err
	}
	return runScenariosWithBases(ctx, r.deps, scenarios, bases)
}
