package runner

import (
	"context"

	"gitea.gitea/mirror/kube-agents-test/pkg/agent"
	"gitea.gitea/mirror/kube-agents-test/pkg/cluster"
	"gitea.gitea/mirror/kube-agents-test/pkg/diag"
	"gitea.gitea/mirror/kube-agents-test/pkg/engine"
)

// Options configures a test suite run.
type Options struct {
	Scenarios []string
	Cluster   cluster.Config
	AgentMode agent.DeploymentMode
	Registry  agent.Registry
	Providers Subsystems
	FailFast  bool
	Parallel  int
}

// Subsystems holds injectable implementations for the test runner.
type Subsystems struct {
	Cluster     cluster.Provider
	Agents      agent.Manager
	Engine      engine.Engine
	Diagnostics diag.Collector
}

// Summary aggregates pass, fail, and error counts.
type Summary struct {
	Passed  int
	Failed  int
	Errored int
}

// SuiteResult is the outcome of running all scenarios in Options.
type SuiteResult struct {
	Results []engine.Result
	Summary Summary
}

// Runner orchestrates cluster, agent, and engine lifecycle for a suite.
type Runner interface {
	Run(ctx context.Context, opts Options) (SuiteResult, error)
}
