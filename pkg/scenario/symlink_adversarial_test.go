package scenario_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadAcceptsSymlinkManifestWithinScenarioDir(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	if err := os.Mkdir(fixtures, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(fixtures, "setup.yaml")
	if err := os.WriteFile(target, []byte("apiVersion: v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "link.yaml")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks not supported in test environment")
	}

	content := []byte(`name: symlink-test
agents:
  - a
setup:
  manifests:
    - link.yaml
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: p
        namespace: test
      conditions:
        - path: .metadata.name
          value: p
  timeout: 30s
`)
	scenarioPath := filepath.Join(root, "scenario.yaml")
	if err := os.WriteFile(scenarioPath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := scenario.Load(scenarioPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if s.Setup.Manifests[0] != "link.yaml" {
		t.Errorf("manifest = %q", s.Setup.Manifests[0])
	}
}

func TestValidateWithSkipsNamespaceChecksWhenSandboxEmpty(t *testing.T) {
	s := validScenario()
	s.Expect.Assertions[0].Resource.Namespace = "kube-system"

	err := s.ValidateWith(scenario.ValidationContext{
		SandboxNamespace: "",
	})
	if err != nil {
		t.Fatalf("ValidateWith empty sandbox unexpectedly failed: %v", err)
	}
}
