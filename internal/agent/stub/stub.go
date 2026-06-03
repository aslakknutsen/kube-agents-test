package stub

import (
	"context"
	"errors"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
)

// ErrNotImplemented indicates the stub manager has no backend yet.
var ErrNotImplemented = errors.New("not implemented")

// Manager is a placeholder agent.Manager for compile-time interface checks.
type Manager struct{}

var _ agent.Manager = (*Manager)(nil)

func (m *Manager) DeploySet(ctx context.Context, agents []string) error {
	return ErrNotImplemented
}

func (m *Manager) Start(ctx context.Context, name string) error {
	return ErrNotImplemented
}

func (m *Manager) Stop(ctx context.Context, name string) error {
	return ErrNotImplemented
}

func (m *Manager) Kill(ctx context.Context, name string) error {
	return ErrNotImplemented
}

func (m *Manager) ApplyDegradation(ctx context.Context, deg agent.Degradation) error {
	return ErrNotImplemented
}

func (m *Manager) ClearDegradation(ctx context.Context, deg agent.Degradation) error {
	return ErrNotImplemented
}

func (m *Manager) Teardown(ctx context.Context) error {
	return ErrNotImplemented
}

// MemoryRegistry resolves agents from an in-memory map (for tests).
type MemoryRegistry struct {
	Specs map[string]agent.Spec
}

var _ agent.AgentRegistry = (*MemoryRegistry)(nil)

// Resolve returns the spec for name or an error if missing.
func (r *MemoryRegistry) Resolve(name string) (agent.Spec, error) {
	if r.Specs == nil {
		return agent.Spec{}, errors.New("agent registry is empty")
	}
	spec, ok := r.Specs[name]
	if !ok {
		return agent.Spec{}, errors.New("unknown agent: " + name)
	}
	return spec, nil
}
