package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// Manager is a stub agent manager until roadmap step 2 is implemented.
type Manager struct {
	Cluster *cluster.Cluster
	Config  agent.Config
}

// New returns a stub manager bound to cluster and config.
func New(c *cluster.Cluster, cfg agent.Config) *Manager {
	return &Manager{Cluster: c, Config: cfg}
}

func (m *Manager) DeploySet(ctx context.Context, agents []string) error {
	_ = ctx
	_ = agents
	return agent.ErrNotImplemented
}

func (m *Manager) Start(ctx context.Context, name string) error {
	_ = ctx
	_ = name
	return agent.ErrNotImplemented
}

func (m *Manager) Stop(ctx context.Context, name string) error {
	_ = ctx
	_ = name
	return agent.ErrNotImplemented
}

func (m *Manager) Kill(ctx context.Context, name string) error {
	_ = ctx
	_ = name
	return agent.ErrNotImplemented
}

func (m *Manager) ApplyDegraded(ctx context.Context, name string, opts agent.DegradedOptions) error {
	_ = ctx
	_ = name
	_ = opts
	return agent.ErrNotImplemented
}

func (m *Manager) Teardown(ctx context.Context) error {
	_ = ctx
	return agent.ErrNotImplemented
}
