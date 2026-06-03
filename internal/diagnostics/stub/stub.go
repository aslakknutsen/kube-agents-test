package stub

import (
	"context"
	"errors"

	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
)

// ErrNotImplemented indicates the stub collector has no backend yet.
var ErrNotImplemented = errors.New("not implemented")

// Collector is a placeholder diagnostics.Collector.
type Collector struct{}

var _ diagnostics.Collector = (*Collector)(nil)

// Collect returns ErrNotImplemented.
func (c *Collector) Collect(ctx context.Context, req diagnostics.CollectRequest) (*diagnostics.Bundle, error) {
	return nil, ErrNotImplemented
}
