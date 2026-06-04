package scenario

import (
	"context"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// Engine executes a single scenario end to end (setup, trigger, poll expect).
//
// API v0 — unstable until first release.
type Engine interface {
	Run(ctx context.Context, sc *Scenario, opts RunOptions) (*Result, error)
}

// RunOptions supplies cluster, agents, and path context for one run.
type RunOptions struct {
	Cluster *cluster.Cluster
	Agents  agent.Manager
	BaseDir string // directory of the scenario file for manifest path resolution
}

// Result is the outcome of one scenario execution.
type Result struct {
	ScenarioName string
	Passed       bool
	Duration     time.Duration
	Failure      *AssertionFailure
}

// AssertionFailure describes why expectations did not converge.
type AssertionFailure struct {
	Message    string
	Conditions []ConditionResult
	TimedOut   bool
}

// ConditionResult compares expected and observed values for one JSONPath.
type ConditionResult struct {
	Resource ObjectRef
	Path     string
	Expected any
	Actual   any
	Matched  bool
}
