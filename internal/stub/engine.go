package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/internal/errs"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

// Engine is a compile-only scenario engine stub.
type Engine struct{}

// NewEngine returns a stub scenario engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Run returns ErrNotImplemented.
func (e *Engine) Run(ctx context.Context, in engine.RunInput) (engine.Result, error) {
	return engine.Result{ScenarioName: in.Scenario.Name}, errs.ErrNotImplemented
}
