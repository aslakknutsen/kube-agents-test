package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	yaml := `
name: test-scenario
description: A simple test scenario
agents:
  - agent-a
  - agent-b
setup:
  manifests:
    - fixtures/ns.yaml
trigger:
  patch:
    apiVersion: apps/v1
    kind: Deployment
    name: target
    namespace: default
    spec:
      replicas: 3
expect:
  - resource:
      apiVersion: apps/v1
      kind: Deployment
      name: target
      namespace: default
    conditions:
      - path: .spec.replicas
        value: 3
timeout: 60s
`
	dir := t.TempDir()
	path := filepath.Join(dir, "scenario.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if s.Name != "test-scenario" {
		t.Errorf("Name = %q, want %q", s.Name, "test-scenario")
	}
	if len(s.Agents) != 2 {
		t.Errorf("len(Agents) = %d, want 2", len(s.Agents))
	}
	if s.Agents[0] != "agent-a" {
		t.Errorf("Agents[0] = %q, want %q", s.Agents[0], "agent-a")
	}
	if len(s.Setup.Manifests) != 1 {
		t.Errorf("len(Setup.Manifests) = %d, want 1", len(s.Setup.Manifests))
	}
	if s.Trigger == nil {
		t.Fatal("Trigger is nil")
	}
	if s.Trigger.Patch.Kind != "Deployment" {
		t.Errorf("Trigger.Patch.Kind = %q, want %q", s.Trigger.Patch.Kind, "Deployment")
	}
	if len(s.Expect) != 1 {
		t.Errorf("len(Expect) = %d, want 1", len(s.Expect))
	}
	if s.Timeout.Seconds() != 60 {
		t.Errorf("Timeout = %v, want 60s", s.Timeout.Duration)
	}
}

func TestLoad_MissingName(t *testing.T) {
	yaml := `
agents:
  - agent-a
`
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestLoad_MissingAgents(t *testing.T) {
	yaml := `
name: no-agents
`
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing agents, got nil")
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()

	scenario1 := `
name: scenario-one
agents:
  - agent-a
expect: []
timeout: 30s
`
	scenario2 := `
name: scenario-two
agents:
  - agent-b
expect: []
timeout: 30s
`
	os.WriteFile(filepath.Join(dir, "one.yaml"), []byte(scenario1), 0644)
	os.WriteFile(filepath.Join(dir, "two.yml"), []byte(scenario2), 0644)
	os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("not a scenario"), 0644)

	scenarios, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir() returned error: %v", err)
	}
	if len(scenarios) != 2 {
		t.Errorf("len(scenarios) = %d, want 2", len(scenarios))
	}
}
