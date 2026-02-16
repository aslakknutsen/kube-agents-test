package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// KindProvider creates ephemeral clusters using kind (Kubernetes IN Docker).
type KindProvider struct {
	// ClusterName is the name passed to `kind create cluster --name`.
	// Defaults to "kube-agents-test" if empty.
	ClusterName string

	kubeconfigPath string
}

func (k *KindProvider) name() string {
	if k.ClusterName != "" {
		return k.ClusterName
	}
	return "kube-agents-test"
}

func (k *KindProvider) Create(ctx context.Context) (string, error) {
	dir, err := os.MkdirTemp("", "kube-agents-test-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir for kubeconfig: %w", err)
	}
	k.kubeconfigPath = filepath.Join(dir, "kubeconfig")

	cmd := exec.CommandContext(ctx,
		"kind", "create", "cluster",
		"--name", k.name(),
		"--kubeconfig", k.kubeconfigPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("kind create cluster: %w", err)
	}

	return k.kubeconfigPath, nil
}

func (k *KindProvider) Destroy(ctx context.Context) error {
	cmd := exec.CommandContext(ctx,
		"kind", "delete", "cluster",
		"--name", k.name(),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kind delete cluster: %w", err)
	}

	if k.kubeconfigPath != "" {
		os.RemoveAll(filepath.Dir(k.kubeconfigPath))
	}
	return nil
}

func (k *KindProvider) Kubeconfig() (string, error) {
	if k.kubeconfigPath == "" {
		return "", fmt.Errorf("no cluster running")
	}
	return k.kubeconfigPath, nil
}
