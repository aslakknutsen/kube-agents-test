package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/internal/errs"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Manager is a compile-only agent manager stub.
type Manager struct{}

// NewManager returns a stub agent manager.
func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Deploy(ctx context.Context, c cluster.Cluster, agents scenario.AgentSet) error {
	return errs.ErrNotImplemented
}

func (m *Manager) Start(ctx context.Context, agentID string) error {
	return errs.ErrNotImplemented
}

func (m *Manager) Stop(ctx context.Context, agentID string) error {
	return errs.ErrNotImplemented
}

func (m *Manager) Kill(ctx context.Context, agentID string) error {
	return errs.ErrNotImplemented
}

func (m *Manager) Teardown(ctx context.Context) error {
	return errs.ErrNotImplemented
}

func (m *Manager) SetResourceLimits(ctx context.Context, agentID string, limits agent.ResourceLimits) error {
	return errs.ErrNotImplemented
}

func (m *Manager) ClearResourceLimits(ctx context.Context, agentID string) error {
	return errs.ErrNotImplemented
}

func (m *Manager) ApplyNetworkPolicy(ctx context.Context, spec agent.NetworkPolicySpec) (string, error) {
	return "", errs.ErrNotImplemented
}

func (m *Manager) RemoveNetworkPolicy(ctx context.Context, policyID string) error {
	return errs.ErrNotImplemented
}
