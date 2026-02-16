package cluster

import (
	"context"
	"fmt"
	"os"
)

// ExistingProvider points at an already-running cluster via a kubeconfig.
// It does not create or destroy anything — useful for dev/staging testing.
type ExistingProvider struct {
	KubeconfigPath string
}

func (e *ExistingProvider) Create(_ context.Context) (string, error) {
	if _, err := os.Stat(e.KubeconfigPath); err != nil {
		return "", fmt.Errorf("kubeconfig not found at %s: %w", e.KubeconfigPath, err)
	}
	return e.KubeconfigPath, nil
}

func (e *ExistingProvider) Destroy(_ context.Context) error {
	// Nothing to tear down for an existing cluster.
	return nil
}

func (e *ExistingProvider) Kubeconfig() (string, error) {
	return e.KubeconfigPath, nil
}
