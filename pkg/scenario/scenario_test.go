package scenario_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadGoldenScenario(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "scenarios", "scaling-agent-respects-quota-agent.yaml")
	s, err := scenario.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if s.Name != "scaling-agent-respects-quota-agent" {
		t.Errorf("Name = %q, want scaling-agent-respects-quota-agent", s.Name)
	}
	if len(s.Agents) != 2 || s.Agents[0] != "scaling-agent" || s.Agents[1] != "quota-agent" {
		t.Errorf("Agents = %v, want [scaling-agent quota-agent]", s.Agents)
	}
	if len(s.Setup.Manifests) != 2 {
		t.Fatalf("Setup.Manifests len = %d, want 2", len(s.Setup.Manifests))
	}
	if s.Trigger == nil || s.Trigger.Patch == nil {
		t.Fatal("expected patch trigger")
	}
	patch := s.Trigger.Patch
	if patch.APIVersion != "apps/v1" || patch.Kind != "Deployment" || patch.Name != "target" || patch.Namespace != "test" {
		t.Errorf("patch identity = %+v", patch)
	}
	replicas, ok := patch.Raw["spec"].(map[string]any)
	if !ok {
		t.Fatalf("patch spec = %#v, want map", patch.Raw["spec"])
	}
	if replicas["replicas"] != 10 {
		t.Errorf("patch replicas = %v, want 10", replicas["replicas"])
	}

	if len(s.Expect.Assertions) != 1 {
		t.Fatalf("Expect.Assertions len = %d, want 1", len(s.Expect.Assertions))
	}
	assertion := s.Expect.Assertions[0]
	if assertion.Resource.Kind != "Deployment" || assertion.Resource.Namespace != "test" {
		t.Errorf("assertion resource = %+v", assertion.Resource)
	}
	if len(assertion.Conditions) != 2 {
		t.Fatalf("conditions len = %d, want 2", len(assertion.Conditions))
	}
	if assertion.Conditions[0].Path != ".spec.replicas" {
		t.Errorf("first condition path = %q", assertion.Conditions[0].Path)
	}
	if s.Expect.Timeout != 120*time.Second {
		t.Errorf("Expect.Timeout = %v, want 120s", s.Expect.Timeout)
	}
}

func TestValidateManifestPathsRejectsTraversal(t *testing.T) {
	err := scenario.ValidateManifestPaths("/tmp/scenarios", []string{"../secret.yaml"})
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestValidateRejectsMultipleTriggers(t *testing.T) {
	s := &scenario.Scenario{
		Name:   "test",
		Agents: scenario.AgentSet{"a"},
		Setup:  scenario.Setup{Manifests: []string{"a.yaml"}},
		Expect: scenario.Expect{
			Assertions: []scenario.ResourceAssertion{{
				Resource: scenario.ResourceSelector{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "p",
					Namespace:  "test",
				},
				Conditions: []scenario.Condition{{Path: ".metadata.name"}},
			}},
			Timeout: time.Minute,
		},
		Trigger: &scenario.Trigger{
			Patch:        &scenario.PatchTrigger{APIVersion: "v1", Kind: "Pod", Name: "p"},
			AgentRestart: &scenario.AgentRestartTrigger{Agent: "a"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected validation error for multiple triggers")
	}
}
