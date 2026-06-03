// Package engine defines the Scenario Engine API for executing a single scenario.
// See docs/scenarios.md and docs/architecture.md#scenario-engine.
package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// ErrTimeout indicates assertion polling exceeded the scenario timeout.
var ErrTimeout = errors.New("assertion timeout")

// InfrastructureError marks failures in cluster, agent, setup, or trigger phases.
type InfrastructureError struct {
	Phase string
	Err   error
}

func (e *InfrastructureError) Error() string {
	return fmt.Sprintf("infrastructure error in %s: %v", e.Phase, e.Err)
}

func (e *InfrastructureError) Unwrap() error { return e.Err }

// Status is the outcome of a scenario execution.
type Status string

const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed" // assertion timeout or mismatch
	StatusError  Status = "error"  // infrastructure/setup failure
)

// ExecuteRequest supplies inputs for a single scenario run.
type ExecuteRequest struct {
	Scenario  *scenario.Scenario
	BaseDir   string // scenario file dir for manifest paths
	Cluster   cluster.Handle
	Agents    agent.Manager
	Collector diagnostics.Collector
}

// Result reports scenario execution outcome.
type Result struct {
	ScenarioName string
	Status       Status
	Duration     time.Duration
	Failure      *Failure
	Err          error // unexpected/internal failures
}

// Failure captures structured assertion failure details.
type Failure struct {
	Reason      string
	Mismatches  []diagnostics.AssertionMismatch
	Diagnostics *diagnostics.Bundle
}

// Engine executes one scenario end-to-end.
type Engine interface {
	Execute(ctx context.Context, req ExecuteRequest) (*Result, error)
}
