package cluster

import "context"

// Provider manages the lifecycle of a Kubernetes cluster for testing.
// Implementations handle kind clusters, k3d, or connecting to an existing cluster.
type Provider interface {
	// Create provisions a new cluster and returns a kubeconfig path.
	Create(ctx context.Context) (kubeconfigPath string, err error)

	// Destroy tears down the cluster. No-op for existing clusters.
	Destroy(ctx context.Context) error

	// Kubeconfig returns the path to the kubeconfig for the active cluster.
	Kubeconfig() string
}

// KindProvider creates ephemeral clusters using kind.
type KindProvider struct {
	ClusterName    string
	KubeconfigPath string
}

func NewKindProvider(clusterName string) *KindProvider {
	return &KindProvider{ClusterName: clusterName}
}

func (k *KindProvider) Create(ctx context.Context) (string, error) {
	// TODO: shell out to `kind create cluster` or use kind's Go library
	return k.KubeconfigPath, nil
}

func (k *KindProvider) Destroy(ctx context.Context) error {
	// TODO: shell out to `kind delete cluster`
	return nil
}

func (k *KindProvider) Kubeconfig() string {
	return k.KubeconfigPath
}

// ExistingClusterProvider points at an already-running cluster.
// Destroy is a no-op.
type ExistingClusterProvider struct {
	KubeconfigPath string
}

func NewExistingClusterProvider(kubeconfigPath string) *ExistingClusterProvider {
	return &ExistingClusterProvider{KubeconfigPath: kubeconfigPath}
}

func (e *ExistingClusterProvider) Create(ctx context.Context) (string, error) {
	return e.KubeconfigPath, nil
}

func (e *ExistingClusterProvider) Destroy(ctx context.Context) error {
	return nil
}

func (e *ExistingClusterProvider) Kubeconfig() string {
	return e.KubeconfigPath
}
