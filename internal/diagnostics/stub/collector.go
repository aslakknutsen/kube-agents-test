package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
)

// Collector is a stub diagnostics collector until roadmap step 4 is implemented.
type Collector struct{}

// New returns a stub collector.
func New() *Collector {
	return &Collector{}
}

// Collect gathers failure diagnostics (not yet implemented).
func (c *Collector) Collect(ctx context.Context, req diagnostics.CollectRequest) (*diagnostics.Bundle, error) {
	_ = ctx
	_ = req
	return nil, diagnostics.ErrNotImplemented
}
