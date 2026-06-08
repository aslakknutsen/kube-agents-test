package scenario_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
	"gopkg.in/yaml.v3"
)

func TestExpectUnmarshalRejectsUnsupportedNodeKind(t *testing.T) {
	var e scenario.Expect
	err := e.UnmarshalYAML(&yaml.Node{Kind: yaml.ScalarNode, Value: "nope"})
	if err == nil {
		t.Fatal("expected error for scalar expect node")
	}
}

func TestExpectUnmarshalSequenceWithoutAssertionsFailsValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "timeout-only.yaml")
	content := []byte(`name: timeout-only
agents:
  - a
setup:
  manifests:
    - f.yaml
expect:
  timeout: 30s
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := scenario.Load(path)
	if err == nil {
		t.Fatal("Load() should fail when expect has timeout but no assertions")
	}
	_ = s
}

func TestExpectUnmarshalSequenceItemMustBeMapping(t *testing.T) {
	var e scenario.Expect
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "not-a-mapping"},
		},
	}
	if err := e.UnmarshalYAML(node); err == nil {
		t.Fatal("expected error for non-mapping sequence item")
	}
}

func TestExpectUnmarshalMappingWithInvalidTimeout(t *testing.T) {
	var e scenario.Expect
	data := []byte(`timeout: not-a-duration
assertions:
  - resource:
      apiVersion: v1
      kind: Pod
      name: p
      namespace: ns
    conditions:
      - path: .metadata.name
        value: p
`)
	if err := yaml.Unmarshal(data, &e); err == nil {
		t.Fatal("expected error for invalid timeout duration")
	}
}

func TestExpectHybridBlockNormalizesTimeoutSibling(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hybrid.yaml")
	content := []byte(`name: hybrid-expect
agents:
  - a
setup:
  manifests:
    - f.yaml
expect:
  - resource:
      apiVersion: v1
      kind: Pod
      name: p
      namespace: test
    conditions:
      - path: .metadata.name
        value: p
  timeout: 45s
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := scenario.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if s.Expect.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want 45s", s.Expect.Timeout)
	}
	if len(s.Expect.Assertions) != 1 {
		t.Fatalf("Assertions len = %d, want 1", len(s.Expect.Assertions))
	}
}
