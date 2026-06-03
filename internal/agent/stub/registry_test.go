package stub_test

import (
	"testing"

	"github.com/kube-agents/kube-agents-test/internal/agent/stub"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
)

func TestMemoryRegistryResolve(t *testing.T) {
	reg := &stub.MemoryRegistry{
		Specs: map[string]agent.Spec{
			"scaling-agent": {Name: "scaling-agent", Image: "example/scaling:latest"},
		},
	}
	spec, err := reg.Resolve("scaling-agent")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Image != "example/scaling:latest" {
		t.Fatalf("spec = %+v", spec)
	}
}

func TestMemoryRegistryUnknownAgent(t *testing.T) {
	reg := &stub.MemoryRegistry{Specs: map[string]agent.Spec{}}
	if _, err := reg.Resolve("missing"); err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestMemoryRegistryNilSpecs(t *testing.T) {
	reg := &stub.MemoryRegistry{}
	if _, err := reg.Resolve("any"); err == nil {
		t.Fatal("expected error for nil specs map")
	}
}
