package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

// Collector is a no-op diagnostics collector stub.
type Collector struct{}

// NewCollector returns a stub diagnostics collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Collect returns empty artifacts without error so failure paths remain testable.
func (c *Collector) Collect(ctx context.Context, fc engine.FailureContext) (diagnostics.Artifacts, error) {
	return diagnostics.Artifacts{ScenarioName: fc.Scenario.Name}, nil
}
