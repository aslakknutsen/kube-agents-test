package stub

import (
	"context"
	"errors"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// ErrNotImplemented indicates the stub provider has no backend yet.
var ErrNotImplemented = errors.New("not implemented")

// Provider is a placeholder cluster.Provider for compile-time interface checks.
type Provider struct{}

var _ cluster.Provider = (*Provider)(nil)

// Acquire returns ErrNotImplemented.
func (p *Provider) Acquire(ctx context.Context) (cluster.Handle, error) {
	return nil, ErrNotImplemented
}

// Handle is a placeholder cluster.Handle.
type Handle struct{}

var _ cluster.Handle = (*Handle)(nil)

// Kubeconfig returns ErrNotImplemented.
func (h *Handle) Kubeconfig() ([]byte, error) {
	return nil, ErrNotImplemented
}

// Release returns ErrNotImplemented.
func (h *Handle) Release(ctx context.Context) error {
	return ErrNotImplemented
}
