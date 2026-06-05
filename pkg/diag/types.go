package diag

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	"gitea.gitea/mirror/kube-agents-test/pkg/cluster"
	"gitea.gitea/mirror/kube-agents-test/pkg/scenario"
)

// Bundle holds artifacts collected on scenario failure.
type Bundle struct {
	AgentLogs        []LogSlice
	Events           []corev1.Event
	Diffs            []ResourceDiff
	MutationTimeline []MutationRecord
}

// LogSlice is filtered log output for one agent or workload.
type LogSlice struct {
	AgentID   string
	Namespace string
	Lines     []string
}

// ResourceDiff compares expected and actual resource state at failure time.
type ResourceDiff struct {
	Resource scenario.ResourceRef
	Expected interface{}
	Actual   interface{}
	Path     string
}

// MutationRecord is one watch event in chronological order.
type MutationRecord struct {
	Timestamp  time.Time
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Operation  string
}

// FailedAssertion describes one condition that did not converge.
type FailedAssertion struct {
	Resource scenario.ResourceRef
	Path     string
	Expected interface{}
	Actual   interface{}
}

// CollectRequest carries context for diagnostics collection.
type CollectRequest struct {
	Scenario         *scenario.Scenario
	FailedAssertions []FailedAssertion
	Cluster          *cluster.Handle
	AgentIDs         []string
	Namespaces       []string
	MutationTimeline []MutationRecord
}

// Collector gathers diagnostics on scenario failure.
type Collector interface {
	Collect(ctx context.Context, req CollectRequest) (*Bundle, error)
}

// NoopCollector returns nil bundle without error.
type NoopCollector struct{}

// Collect implements Collector.
func (NoopCollector) Collect(context.Context, CollectRequest) (*Bundle, error) {
	return nil, nil
}
