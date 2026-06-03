package scenario_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadRejectsEmptyAgents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	yamlDoc := `name: no-agents
agents: []
setup:
  manifests:
    - m.yaml
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: p
        namespace: ns
      conditions:
        - path: .metadata.name
          value: p
  timeout: 10s
`
	if err := os.WriteFile(path, []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Load(path); err == nil {
		t.Fatal("expected validation error for empty agents")
	}
}

func TestLoadRejectsMissingManifests(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	yamlDoc := `name: no-manifests
agents:
  - a
setup:
  manifests: []
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: p
        namespace: ns
      conditions:
        - path: .metadata.name
          value: p
  timeout: 10s
`
	if err := os.WriteFile(path, []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Load(path); err == nil {
		t.Fatal("expected validation error for empty manifests")
	}
}

func TestLoadRejectsInvalidTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	yamlDoc := `name: bad-timeout
agents:
  - a
setup:
  manifests:
    - m.yaml
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: p
        namespace: ns
      conditions:
        - path: .metadata.name
          value: p
  timeout: not-a-duration
`
	if err := os.WriteFile(path, []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := scenario.Load(path); err == nil {
		t.Fatal("expected parse error for invalid timeout")
	}
}

func TestLoadLegacyTimeoutBeforeAssertionsFails(t *testing.T) {
	// Legacy rewriter only handles list-then-timeout order (README shape).
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	yamlDoc := `name: timeout-first
agents:
  - a
setup:
  manifests:
    - m.yaml
expect:
  timeout: 10s
  - resource:
      apiVersion: v1
      kind: Pod
      name: p
      namespace: ns
    conditions:
      - path: .metadata.name
        value: p
`
	if err := os.WriteFile(path, []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := scenario.Load(path)
	if err == nil {
		t.Fatal("expected error when timeout precedes assertion list in legacy expect block")
	}
}

func TestLoadTriggerFault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fault.yaml")
	yamlDoc := `name: fault-trigger
agents:
  - scaling-agent
setup:
  manifests:
    - m.yaml
trigger:
  fault:
    kind: killAgent
    agent: scaling-agent
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: p
        namespace: ns
      conditions:
        - path: .metadata.name
          value: p
  timeout: 30s
`
	if err := os.WriteFile(path, []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := scenario.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Trigger == nil || s.Trigger.Fault == nil {
		t.Fatal("expected trigger.fault")
	}
	if s.Trigger.Fault.Kind != "killAgent" || s.Trigger.Fault.Agent != "scaling-agent" {
		t.Fatalf("fault = %+v", s.Trigger.Fault)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := scenario.Load("/nonexistent/scenario.yaml"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestResolveManifestPathsRelativeAndAbsolute(t *testing.T) {
	s := &scenario.Scenario{
		Setup: scenario.Setup{Manifests: []string{"rel.yaml", "/abs.yaml"}},
	}
	paths, err := s.ResolveManifestPaths("/base/dir")
	if err != nil {
		t.Fatal(err)
	}
	if paths[0] != filepath.Join("/base/dir", "rel.yaml") {
		t.Errorf("relative path = %q", paths[0])
	}
	if paths[1] != "/abs.yaml" {
		t.Errorf("absolute path = %q", paths[1])
	}
}

func TestValidateTriggerPatchMissingFields(t *testing.T) {
	s := &scenario.Scenario{
		Name:   "x",
		Agents: []string{"a"},
		Setup:  scenario.Setup{Manifests: []string{"m.yaml"}},
		Expect: scenario.Expect{
			Timeout: 10 * time.Second,
			Assertions: []scenario.ResourceAssertion{{
				Resource: scenario.ResourceRef{
					APIVersion: "v1", Kind: "Pod", Name: "p", Namespace: "ns",
				},
				Conditions: []scenario.PathCondition{{Path: ".x", Value: 1}},
			}},
		},
		Trigger: &scenario.Trigger{
			Patch: &scenario.ResourcePatch{
				Kind: "Deployment", Name: "d", Namespace: "ns",
			},
		},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected validation error for incomplete trigger.patch")
	}
}

