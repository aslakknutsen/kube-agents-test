package diagnostics

import (
	"context"
)

// Report contains failure diagnostics for a scenario.
type Report struct {
	AgentLogs     map[string]string // agent name -> logs
	Events        string            // kubernetes events in the test namespace
	ResourceDiffs []ResourceDiff    // expected vs actual
	Timeline      []MutationEvent   // recorded watch events
}

// ResourceDiff shows the difference between expected and actual state.
type ResourceDiff struct {
	Resource string
	Expected string
	Actual   string
}

// MutationEvent is a single resource change observed during the test.
type MutationEvent struct {
	Timestamp string
	Resource  string
	Action    string // ADDED, MODIFIED, DELETED
	Summary   string
}

// Collector gathers diagnostics when a scenario fails.
type Collector struct {
	Kubeconfig string
}

func NewCollector(kubeconfig string) *Collector {
	return &Collector{Kubeconfig: kubeconfig}
}

// Collect gathers agent logs, events, resource diffs, and watch timeline
// for the given agents and namespace.
func (c *Collector) Collect(ctx context.Context, agentNames []string, namespace string) *Report {
	report := &Report{
		AgentLogs: make(map[string]string),
	}

	for _, agentName := range agentNames {
		report.AgentLogs[agentName] = c.fetchAgentLogs(ctx, agentName)
	}

	report.Events = c.fetchEvents(ctx, namespace)

	// TODO: compute resource diffs and timeline
	return report
}

func (c *Collector) fetchAgentLogs(ctx context.Context, agentName string) string {
	// TODO: fetch logs from agent pods via client-go
	return ""
}

func (c *Collector) fetchEvents(ctx context.Context, namespace string) string {
	// TODO: list events in the namespace
	return ""
}
