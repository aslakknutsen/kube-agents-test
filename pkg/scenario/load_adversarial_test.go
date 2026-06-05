package scenario_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadRejectsTraversalManifestInFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte(`name: traversal-test
agents:
  - agent-a
setup:
  manifests:
    - ../../../etc/passwd
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
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := scenario.Load(path)
	if err == nil {
		t.Fatal("Load() should reject manifest path traversal")
	}
	if !strings.Contains(err.Error(), "..") && !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("error = %q, want traversal rejection", err)
	}
}

func TestLoadDirEmptyDirectoryFails(t *testing.T) {
	dir := t.TempDir()
	_, err := scenario.LoadDir(dir)
	if err == nil {
		t.Fatal("LoadDir() should fail on empty directory")
	}
	if !strings.Contains(err.Error(), "no scenario files") {
		t.Fatalf("error = %q", err)
	}
}

func TestLoadPathsEmptySliceFails(t *testing.T) {
	_, err := scenario.LoadPaths(nil)
	if err == nil {
		t.Fatal("LoadPaths() should fail on empty paths")
	}
}

func TestLoadRejectsMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.yaml")
	if err := os.WriteFile(path, []byte("name: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := scenario.Load(path)
	if err == nil {
		t.Fatal("Load() should fail on malformed YAML")
	}
}

func TestLoadRejectsMissingRequiredFields(t *testing.T) {
	dir := t.TempDir()
	content := []byte(`name: incomplete
agents: []
setup:
  manifests: []
expect:
  assertions: []
  timeout: 0s
`)
	path := filepath.Join(dir, "incomplete.yaml")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := scenario.Load(path)
	if err == nil {
		t.Fatal("Load() should fail validation for incomplete scenario")
	}
}
