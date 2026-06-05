package scenario_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitea.gitea/mirror/kube-agents-test/pkg/fault"
	"gitea.gitea/mirror/kube-agents-test/pkg/scenario"
	"gopkg.in/yaml.v3"
)

const docExampleYAML = `name: scaling-agent-respects-quota-agent
description: >
  When the scaling agent wants to add replicas but the quota agent has
  capped the namespace, the deployment should stay at the capped count.

agents:
  - scaling-agent
  - quota-agent

setup:
  manifests:
    - fixtures/namespace-with-quota.yaml
    - fixtures/deployment-at-limit.yaml

trigger:
  patch:
    apiVersion: apps/v1
    kind: Deployment
    name: target
    namespace: test
    spec:
      replicas: 10

expect:
  timeout: 120s
  assertions:
    - resource:
        apiVersion: apps/v1
        kind: Deployment
        name: target
        namespace: test
      conditions:
        - path: .spec.replicas
          value: 5
        - path: .status.readyReplicas
          value: 5
`

func TestLoadDocExample(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scaling.yaml")
	if err := os.WriteFile(path, []byte(docExampleYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := scenario.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if s.Name != "scaling-agent-respects-quota-agent" {
		t.Errorf("name: got %q", s.Name)
	}
	if len(s.Agents) != 2 || s.Agents[0] != "scaling-agent" {
		t.Errorf("agents: %#v", s.Agents)
	}
	if len(s.Setup.Manifests) != 2 {
		t.Errorf("manifests: %#v", s.Setup.Manifests)
	}
	if s.Trigger == nil || s.Trigger.Patch == nil {
		t.Fatal("expected patch trigger")
	}
	if s.Trigger.Patch.Name != "target" {
		t.Errorf("patch name: %q", s.Trigger.Patch.Name)
	}
	if s.Expect.Timeout != 120*time.Second {
		t.Errorf("timeout: %v", s.Expect.Timeout)
	}
	if len(s.Expect.Assertions) != 1 {
		t.Fatalf("assertions: %d", len(s.Expect.Assertions))
	}
	a := s.Expect.Assertions[0]
	if len(a.Conditions) != 2 {
		t.Fatalf("conditions: %d", len(a.Conditions))
	}
	if a.Conditions[0].Path != ".spec.replicas" {
		t.Errorf("path: %q", a.Conditions[0].Path)
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	original := &scenario.Scenario{
		Name:        "round-trip",
		Description: "test",
		Agents:      []string{"a"},
		Setup:       scenario.Setup{Manifests: []string{"m.yaml"}},
		Trigger: &fault.Trigger{
			AgentRestart: &fault.AgentRestartTrigger{AgentID: "a"},
		},
		Expect: scenario.Expect{
			Timeout: 30 * time.Second,
			Assertions: []scenario.Assertion{
				{
					Resource: scenario.ResourceRef{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       "p",
						Namespace:  "ns",
					},
					Conditions: []scenario.Condition{
						{Path: ".status.phase", Value: "Running"},
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded scenario.Scenario
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatal(err)
	}
	if decoded.Name != original.Name {
		t.Errorf("name: got %q want %q", decoded.Name, original.Name)
	}
	if decoded.Expect.Timeout != original.Expect.Timeout {
		t.Errorf("timeout: got %v want %v", decoded.Expect.Timeout, original.Expect.Timeout)
	}
}

func TestValidateErrors(t *testing.T) {
	tests := []struct {
		name string
		s    scenario.Scenario
	}{
		{"empty name", scenario.Scenario{Agents: []string{"a"}, Setup: scenario.Setup{Manifests: []string{"m"}}, Expect: scenario.Expect{Timeout: time.Second, Assertions: []scenario.Assertion{{Resource: scenario.ResourceRef{APIVersion: "v1", Kind: "Pod", Name: "p"}, Conditions: []scenario.Condition{{Path: ".x", Value: 1}}}}}}},
		{"no agents", scenario.Scenario{Name: "x", Setup: scenario.Setup{Manifests: []string{"m"}}, Expect: scenario.Expect{Timeout: time.Second, Assertions: []scenario.Assertion{{Resource: scenario.ResourceRef{APIVersion: "v1", Kind: "Pod", Name: "p"}, Conditions: []scenario.Condition{{Path: ".x", Value: 1}}}}}}},
		{"zero timeout", scenario.Scenario{Name: "x", Agents: []string{"a"}, Setup: scenario.Setup{Manifests: []string{"m"}}, Expect: scenario.Expect{Assertions: []scenario.Assertion{{Resource: scenario.ResourceRef{APIVersion: "v1", Kind: "Pod", Name: "p"}, Conditions: []scenario.Condition{{Path: ".x", Value: 1}}}}}}},
		{"whitespace name", scenario.Scenario{Name: "   ", Agents: []string{"a"}, Setup: scenario.Setup{Manifests: []string{"m"}}, Expect: scenario.Expect{Timeout: time.Second, Assertions: []scenario.Assertion{{Resource: scenario.ResourceRef{APIVersion: "v1", Kind: "Pod", Name: "p"}, Conditions: []scenario.Condition{{Path: ".x", Value: 1}}}}}}},
		{"whitespace agent", scenario.Scenario{Name: "x", Agents: []string{"   "}, Setup: scenario.Setup{Manifests: []string{"m"}}, Expect: scenario.Expect{Timeout: time.Second, Assertions: []scenario.Assertion{{Resource: scenario.ResourceRef{APIVersion: "v1", Kind: "Pod", Name: "p"}, Conditions: []scenario.Condition{{Path: ".x", Value: 1}}}}}}},
		{"negative timeout", scenario.Scenario{Name: "x", Agents: []string{"a"}, Setup: scenario.Setup{Manifests: []string{"m"}}, Expect: scenario.Expect{Timeout: -5 * time.Second, Assertions: []scenario.Assertion{{Resource: scenario.ResourceRef{APIVersion: "v1", Kind: "Pod", Name: "p"}, Conditions: []scenario.Condition{{Path: ".x", Value: 1}}}}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.s.Validate()
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, scenario.ErrInvalidScenario) {
				t.Errorf("errors.Is: got %v", err)
			}
		})
	}
}

func TestLegacyExpectSequence(t *testing.T) {
	const legacy = `name: legacy
agents:
  - a
setup:
  manifests:
    - m.yaml
expect:
  - resource:
      apiVersion: v1
      kind: Pod
      name: p
    conditions:
      - path: .status.phase
        value: Running
  - timeout: 60s
`
	var s scenario.Scenario
	if err := yaml.Unmarshal([]byte(legacy), &s); err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if s.Expect.Timeout != 60*time.Second {
		t.Errorf("timeout: %v", s.Expect.Timeout)
	}
}
