package scenario_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadDir_emptyReturnsNoScenarios(t *testing.T) {
	dir := t.TempDir()
	got, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestLoadDir_skipsNonYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestLoadFile_docsMixedExpectShapeIsInvalidYAML(t *testing.T) {
	// docs/scenarios.md shows a list entry and timeout as mapping siblings under expect;
	// that shape does not parse as valid YAML (see handover on issue #14).
	_, err := scenario.LoadFile(filepath.Join("testfixtures", "mixed-expect-docs-shape.yaml"))
	if err == nil {
		t.Fatal("expected parse error for documented mixed expect shape")
	}
	if !strings.Contains(err.Error(), "parse scenario") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadFile_sequenceExpectWithoutTimeoutFailsValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seq.yaml")
	content := `name: seq
agents:
  - a
setup:
  manifests: []
expect:
  - resource:
      apiVersion: v1
      kind: Pod
      name: p
    conditions:
      - path: .metadata.name
        value: p
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := scenario.LoadFile(path)
	if err == nil || !strings.Contains(err.Error(), "expect.timeout") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidate_rejectsWhitespaceAgentName(t *testing.T) {
	sc := &scenario.Scenario{
		Name:   "x",
		Agents: []string{"  "},
		Expect: scenario.Expectation{
			Resources: []scenario.ResourceExpect{{
				Resource:   scenario.ObjectRef{APIVersion: "v1", Kind: "Pod", Name: "p"},
				Conditions: []scenario.Condition{{Path: ".x", Value: 1}},
			}},
			Timeout: scenario.DefaultTimeout,
		},
	}
	if err := sc.Validate(); err == nil || !strings.Contains(err.Error(), "agents[0]") {
		t.Fatalf("err = %v", err)
	}
}

func TestValidate_rejectsPatchMissingAPIVersion(t *testing.T) {
	sc := &scenario.Scenario{
		Name:   "x",
		Agents: []string{"a"},
		Trigger: &scenario.Trigger{
			Patch: &scenario.ResourcePatch{
				ObjectRef: scenario.ObjectRef{Kind: "Pod", Name: "p"},
			},
		},
		Expect: scenario.Expectation{
			Resources: []scenario.ResourceExpect{{
				Resource:   scenario.ObjectRef{APIVersion: "v1", Kind: "Pod", Name: "p"},
				Conditions: []scenario.Condition{{Path: ".x", Value: 1}},
			}},
			Timeout: scenario.DefaultTimeout,
		},
	}
	if err := sc.Validate(); err == nil || !strings.Contains(err.Error(), "trigger.patch.apiVersion") {
		t.Fatalf("err = %v", err)
	}
}
