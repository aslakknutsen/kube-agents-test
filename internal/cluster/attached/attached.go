// Package attached will implement cluster.Provider for existing kubeconfigs.
package attached

import (
	"context"
	"errors"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// ErrNotImplemented indicates attached cluster support is not implemented yet.
var ErrNotImplemented = errors.New("not implemented")

// Provider validates and uses an existing cluster kubeconfig.
type Provider struct {
	Config cluster.AttachedConfig
}

var _ cluster.Provider = (*Provider)(nil)

// Acquire returns ErrNotImplemented.
func (p *Provider) Acquire(ctx context.Context) (cluster.Handle, error) {
	return nil, ErrNotImplemented
}

// Handle holds kubeconfig access for an attached cluster.
type Handle struct{}

var _ cluster.Handle = (*Handle)(nil)

// Kubeconfig returns ErrNotImplemented.
func (h *Handle) Kubeconfig() ([]byte, error) {
	return nil, ErrNotImplemented
}

// Release is a no-op for attached clusters once implemented.
func (h *Handle) Release(ctx context.Context) error {
	return nil
}
