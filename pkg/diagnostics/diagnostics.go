// Package diagnostics defines failure artifact types and collection interfaces.
// See docs/failure-diagnostics.md.
package diagnostics

import (
	"context"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// AssertionMismatch records an expected vs actual field mismatch at failure time.
type AssertionMismatch struct {
	Resource scenario.ResourceRef
	Path     string
	Expected interface{}
	Actual   interface{}
}

// MutationEvent records a resource change observed during scenario execution.
type MutationEvent struct {
	Timestamp time.Time
	Resource  scenario.ResourceRef
	Operation string // e.g. Added, Modified, Deleted
}

// AgentLog holds filtered log output for an agent involved in the scenario.
type AgentLog struct {
	Agent   string
	Content string
}

// Event is a Kubernetes event in the test namespace.
type Event struct {
	Type      string
	Reason    string
	Message   string
	Involved  scenario.ResourceRef
	Timestamp time.Time
}

// FieldDiff compares an expected assertion value to what was observed.
type FieldDiff struct {
	Resource scenario.ResourceRef
	Path     string
	Expected interface{}
	Actual   interface{}
}

// Bundle aggregates diagnostic artifacts collected on assertion failure.
type Bundle struct {
	AgentLogs []AgentLog
	Events    []Event
	Diff      []FieldDiff
	Timeline  []MutationEvent
}

// CollectRequest supplies context for diagnostic collection.
type CollectRequest struct {
	Scenario   *scenario.Scenario
	Mismatches []AssertionMismatch
	WatchLog   []MutationEvent
}

// Collector gathers diagnostic artifacts when assertions fail.
type Collector interface {
	Collect(ctx context.Context, req CollectRequest) (*Bundle, error)
}
