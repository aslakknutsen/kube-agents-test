// Package cluster defines the interface for creating and managing test clusters.
package cluster

import "context"

// Provider manages the lifecycle of a Kubernetes cluster for testing.
type Provider interface {
	// Create provisions a new cluster and returns a kubeconfig path.
	Create(ctx context.Context) (kubeconfigPath string, err error)

	// Destroy tears down the cluster. Must be safe to call multiple times.
	Destroy(ctx context.Context) error

	// Kubeconfig returns the path to the kubeconfig for the current cluster.
	// Returns an error if no cluster is running.
	Kubeconfig() (string, error)
}
