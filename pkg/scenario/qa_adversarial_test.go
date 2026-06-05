package scenario_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestValidateWithRejectsPatchTriggerOutsideSandbox(t *testing.T) {
	s := validScenario()
	s.Trigger = &scenario.Trigger{
		Patch: &scenario.PatchTrigger{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "p",
			Namespace:  "kube-system",
		},
	}

	err := s.ValidateWith(scenario.ValidationContext{
		SandboxNamespace:   "sandbox",
		AllowClusterScoped: false,
	})
	if err == nil {
		t.Fatal("expected patch trigger outside sandbox to be rejected")
	}
	if !strings.Contains(err.Error(), "outside sandbox") {
		t.Fatalf("error = %q", err)
	}
}

func TestTriggerValidateRejectsMultipleArms(t *testing.T) {
	tr := &scenario.Trigger{
		Patch: &scenario.PatchTrigger{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "p",
		},
		Fault: &scenario.FaultTrigger{Type: "kill-agent"},
	}
	err := tr.Validate()
	if err == nil {
		t.Fatal("expected error when multiple trigger arms are set")
	}
	if !strings.Contains(err.Error(), "at most one") {
		t.Fatalf("error = %q", err)
	}
}

func TestValidateWithAllowsClusterScopedWhenFlagSet(t *testing.T) {
	s := validScenario()
	s.Expect.Assertions[0].Resource.Namespace = ""

	err := s.ValidateWith(scenario.ValidationContext{
		SandboxNamespace:   "sandbox",
		AllowClusterScoped: true,
	})
	if err != nil {
		t.Fatalf("ValidateWith() error = %v, want nil when AllowClusterScoped=true", err)
	}
}

func TestValidateRejectsDuplicateAgentIDs(t *testing.T) {
	s := validScenario()
	s.Agents = scenario.AgentSet{"agent-a", "agent-a"}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil (duplicates not forbidden in v1)", err)
	}
}

func TestValidateManifestPathsRejectsLeadingDotDot(t *testing.T) {
	err := scenario.ValidateManifestPaths("/tmp/scenarios", []string{"../secret.yaml"})
	if err == nil {
		t.Fatal("expected error for leading .. segment")
	}
}

func TestExpectHybridBlockWithTabIndentationFailsLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tab-expect.yaml")
	content := []byte("name: tab-hybrid\nagents:\n  - a\nsetup:\n  manifests:\n    - f.yaml\nexpect:\n\t- resource:\n\t    apiVersion: v1\n\t    kind: Pod\n\t    name: p\n\t    namespace: test\n\t  conditions:\n\t    - path: .metadata.name\n\t      value: p\n\ttimeout: 30s\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := scenario.Load(path)
	if err == nil {
		t.Fatal("Load() should fail or mis-parse tab-indented hybrid expect block")
	}
}
