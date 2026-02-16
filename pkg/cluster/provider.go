package cluster

import "context"

// Provider manages the lifecycle of a Kubernetes cluster for testing.
// Implementations handle kind, k3d, or connecting to an existing cluster.
type Provider interface {
	// Create provisions a new cluster and returns a kubeconfig path.
	Create(ctx context.Context) (kubeconfigPath string, err error)

	// Destroy tears down the cluster. No-op for external clusters.
	Destroy(ctx context.Context) error

	// Kubeconfig returns the path to the kubeconfig for the active cluster.
	Kubeconfig() string
}

// KindProvider creates ephemeral clusters using kind.
type KindProvider struct {
	clusterName    string
	kubeconfigPath string
}

func NewKindProvider(clusterName string) *KindProvider {
	return &KindProvider{clusterName: clusterName}
}

func (k *KindProvider) Create(ctx context.Context) (string, error) {
	// TODO: shell out to `kind create cluster` or use kind's Go library
	return k.kubeconfigPath, nil
}

func (k *KindProvider) Destroy(ctx context.Context) error {
	// TODO: shell out to `kind delete cluster`
	return nil
}

func (k *KindProvider) Kubeconfig() string {
	return k.kubeconfigPath
}

// ExistingClusterProvider connects to a pre-existing cluster via kubeconfig.
type ExistingClusterProvider struct {
	kubeconfigPath string
}

func NewExistingClusterProvider(kubeconfigPath string) *ExistingClusterProvider {
	return &ExistingClusterProvider{kubeconfigPath: kubeconfigPath}
}

func (e *ExistingClusterProvider) Create(_ context.Context) (string, error) {
	return e.kubeconfigPath, nil
}

func (e *ExistingClusterProvider) Destroy(_ context.Context) error {
	return nil
}

func (e *ExistingClusterProvider) Kubeconfig() string {
	return e.kubeconfigPath
}
