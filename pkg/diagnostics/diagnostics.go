// Package diagnostics defines the interface and types for collecting
// failure diagnostics from the cluster when a scenario fails.
package diagnostics

import "context"

// Scope defines what to collect diagnostics for.
type Scope struct {
	// Namespace to collect events and resources from.
	Namespace string
	// AgentNames to collect logs for.
	AgentNames []string
}

// Report holds the collected diagnostic data from a failed scenario.
type Report struct {
	// AgentLogs maps agent name to its log output.
	AgentLogs map[string]string

	// Events are the Kubernetes events from the test namespace.
	Events []string

	// ResourceDiffs describes mismatches between expected and actual state.
	ResourceDiffs []ResourceDiff

	// CollectionError is set if diagnostics collection itself failed.
	CollectionError string
}

// ResourceDiff captures the difference between expected and actual values
// for a single resource field.
type ResourceDiff struct {
	Resource string
	Path     string
	Expected interface{}
	Actual   interface{}
}

// Collector gathers diagnostic information from the cluster.
type Collector interface {
	// Collect gathers diagnostics for the given scope.
	Collect(ctx context.Context, scope Scope) (*Report, error)
}
