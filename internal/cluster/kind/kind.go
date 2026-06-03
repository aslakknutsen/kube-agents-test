// Package kind will implement ephemeral cluster provisioning via kind.
// See docs/architecture.md#cluster-provider.
package kind

import (
	"context"
	"errors"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// ErrNotImplemented indicates kind provisioning is not implemented yet.
var ErrNotImplemented = errors.New("not implemented")

// Provider implements cluster.Provider for kind clusters.
type Provider struct {
	Config cluster.EphemeralConfig
}

var _ cluster.Provider = (*Provider)(nil)

// Acquire returns ErrNotImplemented until kind integration lands.
func (p *Provider) Acquire(ctx context.Context) (cluster.Handle, error) {
	return nil, ErrNotImplemented
}

// Handle exposes kubeconfig for a kind cluster.
type Handle struct{}

var _ cluster.Handle = (*Handle)(nil)

// Kubeconfig returns ErrNotImplemented.
func (h *Handle) Kubeconfig() ([]byte, error) {
	return nil, ErrNotImplemented
}

// Release tears down the kind cluster.
func (h *Handle) Release(ctx context.Context) error {
	return ErrNotImplemented
}
