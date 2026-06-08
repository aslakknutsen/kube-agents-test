package diagnostics

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

// Collector gathers supplemental failure artifacts. Invoked only on scenario failure.
type Collector interface {
	Collect(ctx context.Context, fc engine.FailureContext) (Artifacts, error)
}
