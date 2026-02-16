package diagnostics

import "context"

// Report contains failure diagnostics collected after a scenario fails.
type Report struct {
	// AgentLogs maps agent name to its log output during the scenario.
	AgentLogs map[string]string

	// Events contains Kubernetes events from the test namespace.
	Events []string

	// ResourceDiffs contains textual diffs between expected and actual state.
	ResourceDiffs []string
}

// Collector gathers diagnostic information from the cluster after a
// scenario failure.
type Collector struct {
	kubeconfig string
}

func NewCollector(kubeconfig string) *Collector {
	return &Collector{kubeconfig: kubeconfig}
}

// Collect gathers diagnostics for a failed scenario.
func (c *Collector) Collect(ctx context.Context, scenarioName string, agents []string) (*Report, error) {
	report := &Report{
		AgentLogs: make(map[string]string),
	}

	// TODO: fetch agent logs via client-go pod log API
	for _, a := range agents {
		report.AgentLogs[a] = ""
	}

	// TODO: list events in the scenario namespace
	// TODO: compute resource diffs between expected and actual

	return report, nil
}
