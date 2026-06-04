package kind

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// Provider is a stub kind backend; Provision/Teardown return ErrNotImplemented until implemented.
type Provider struct {
	Opts cluster.EphemeralOptions
}

// New returns a kind cluster provider stub.
func New(opts cluster.EphemeralOptions) *Provider {
	if opts.Backend == "" {
		opts.Backend = "kind"
	}
	return &Provider{Opts: opts}
}

// Provision creates an ephemeral kind cluster (not yet implemented).
func (p *Provider) Provision(ctx context.Context) (*cluster.Cluster, error) {
	_ = ctx
	return nil, cluster.ErrNotImplemented
}

// Attach validates and uses an existing kubeconfig (not yet implemented).
func (p *Provider) Attach(ctx context.Context, kubeconfig string) (*cluster.Cluster, error) {
	_ = ctx
	_ = kubeconfig
	return nil, cluster.ErrNotImplemented
}

// Teardown deletes the ephemeral cluster (not yet implemented).
func (p *Provider) Teardown(ctx context.Context) error {
	_ = ctx
	return cluster.ErrNotImplemented
}
