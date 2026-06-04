package cluster_test

import (
	"testing"

	intcluster "github.com/kube-agents/kube-agents-test/internal/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

func TestNewProvider_unknownBackend(t *testing.T) {
	_, err := intcluster.NewProvider(cluster.EphemeralOptions{Backend: "minikube"})
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
}

func TestNewProvider_kindDefault(t *testing.T) {
	p, err := intcluster.NewProvider(cluster.EphemeralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if p == nil {
		t.Fatal("nil provider")
	}
}
