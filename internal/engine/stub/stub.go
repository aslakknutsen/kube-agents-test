package stub

import (
	"context"
	"errors"

	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

// ErrNotImplemented indicates the stub engine has no backend yet.
var ErrNotImplemented = errors.New("not implemented")

// Engine is a placeholder engine.Engine for compile-time interface checks.
type Engine struct{}

var _ engine.Engine = (*Engine)(nil)

// Execute returns a StatusError result with ErrNotImplemented.
func (e *Engine) Execute(ctx context.Context, req engine.ExecuteRequest) (*engine.Result, error) {
	if req.Scenario == nil {
		return nil, errors.New("scenario is required")
	}
	return &engine.Result{
		ScenarioName: req.Scenario.Name,
		Status:       engine.StatusError,
		Err:          ErrNotImplemented,
	}, nil
}
