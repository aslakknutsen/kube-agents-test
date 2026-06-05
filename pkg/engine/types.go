package engine

import (
	"context"
	"time"

	"gitea.gitea/mirror/kube-agents-test/pkg/agent"
	"gitea.gitea/mirror/kube-agents-test/pkg/cluster"
	"gitea.gitea/mirror/kube-agents-test/pkg/diag"
	"gitea.gitea/mirror/kube-agents-test/pkg/scenario"
)

// ResultStatus is the outcome of a single scenario run.
type ResultStatus int

const (
	Pass ResultStatus = iota
	Fail
	Error
)

// FailureReason categorizes why a scenario did not pass.
type FailureReason int

const (
	ReasonTimeout FailureReason = iota
	ReasonAssertionMismatch
	ReasonTriggerFailed
	ReasonSetupFailed
)

// RunInput is everything the scenario engine needs for one run.
type RunInput struct {
	Scenario      *scenario.Scenario
	Cluster       *cluster.Handle
	Agents        agent.Manager
	Diagnostics   diag.Collector
	WorkDir       string
	WatchMutation bool
}

// Result is the outcome of executing one scenario.
type Result struct {
	ScenarioName string
	Status       ResultStatus
	Duration     time.Duration
	Failure      *FailureDetail
}

// FailureDetail is populated on Fail or Error when details are available.
type FailureDetail struct {
	Reason      FailureReason
	Message     string
	Assertions  []AssertionFailure
	Diagnostics *diag.Bundle
}

// AssertionFailure records one condition that did not match at failure time.
type AssertionFailure struct {
	Resource scenario.ResourceRef
	Path     string
	Expected interface{}
	Actual   interface{}
}

// Engine executes a single scenario end to end.
type Engine interface {
	Run(ctx context.Context, in RunInput) Result
}
