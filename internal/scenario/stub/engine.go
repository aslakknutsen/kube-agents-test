package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Engine is a stub scenario engine until roadmap step 3 is implemented.
type Engine struct{}

// New returns a stub scenario engine.
func New() *Engine {
	return &Engine{}
}

// Run applies setup, trigger, and assertions (not yet implemented).
func (e *Engine) Run(ctx context.Context, sc *scenario.Scenario, opts scenario.RunOptions) (*scenario.Result, error) {
	_ = ctx
	_ = sc
	_ = opts
	return nil, scenario.ErrNotImplemented
}

var _ scenario.Engine = (*Engine)(nil)
