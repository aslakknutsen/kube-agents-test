package cluster

import (
	"fmt"

	"github.com/kube-agents/kube-agents-test/internal/cluster/kind"
	pkgcluster "github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// NewProvider returns a Provider for the given ephemeral backend name.
func NewProvider(opts pkgcluster.EphemeralOptions) (pkgcluster.Provider, error) {
	backend := opts.Backend
	if backend == "" {
		backend = "kind"
	}
	switch backend {
	case "kind":
		return kind.New(opts), nil
	default:
		return nil, fmt.Errorf("cluster: unknown backend %q", backend)
	}
}
