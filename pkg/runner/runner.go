// Package runner orchestrates cluster, agent, scenario, and diagnostics subsystems.
//
// API v0 — unstable until first release.
package runner

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Config wires subsystem implementations and run options.
type Config struct {
	ClusterProvider cluster.Provider
	AgentManager    agent.Manager
	ScenarioEngine  scenario.Engine
	Diagnostics     diagnostics.Collector

	ClusterMode   cluster.Mode
	Kubeconfig    string
	EphemeralOpts cluster.EphemeralOptions
	AgentConfig   agent.Config

	// AgentManagerFactory builds a manager after the cluster is ready when AgentManager is nil.
	AgentManagerFactory agent.Factory
}

// Runner executes one or more scenarios and aggregates results.
type Runner interface {
	Run(ctx context.Context, scenarios []*scenario.Scenario) (*SuiteResult, error)
}

// SuiteResult aggregates per-scenario outcomes.
type SuiteResult struct {
	Results []ScenarioRunResult
	Passed  int
	Failed  int
}

// ScenarioRunResult is the outcome of one scenario including infra errors.
type ScenarioRunResult struct {
	Name        string
	Passed      bool
	Result      *scenario.Result
	Diagnostics *diagnostics.Bundle
	Err         error
}
