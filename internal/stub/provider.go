package stub

import (
	"context"

	"github.com/kube-agents/kube-agents-test/internal/errs"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// Provider is a compile-only cluster provider stub.
type Provider struct{}

// NewProvider returns a stub cluster provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Ensure returns ErrNotImplemented.
func (p *Provider) Ensure(ctx context.Context, cfg cluster.Config) (cluster.Cluster, error) {
	return nil, errs.ErrNotImplemented
}

// Teardown returns ErrNotImplemented.
func (p *Provider) Teardown(ctx context.Context, c cluster.Cluster) error {
	return errs.ErrNotImplemented
}
