// Package diagnostics defines failure diagnostics collection for non-converged scenarios.
//
// API v0 — unstable until first release.
package diagnostics

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Bundle holds artifacts collected on scenario failure.
type Bundle struct {
	AgentLogs        map[string][]byte
	Events           []corev1.Event
	ExpectDiff       []FieldDiff
	MutationTimeline []MutationEvent
}

// FieldDiff is one expected-vs-actual field mismatch.
type FieldDiff struct {
	Resource scenario.ObjectRef
	Path     string
	Expected any
	Actual   any
}

// MutationEvent records one observed resource change during the test.
type MutationEvent struct {
	Timestamp string
	Resource  scenario.ObjectRef
	Verb      string
	Detail    string
}

// CollectRequest inputs for diagnostics collection.
type CollectRequest struct {
	Cluster   *cluster.Cluster
	Scenario  *scenario.Scenario
	Failure   *scenario.AssertionFailure
	Namespace string
}

// Collector gathers diagnostics on failure.
type Collector interface {
	Collect(ctx context.Context, req CollectRequest) (*Bundle, error)
}
