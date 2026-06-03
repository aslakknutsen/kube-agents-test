package scenario_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

const readmeExample = `name: scaling-agent-respects-quota-agent
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
  timeout: 120s
`

func TestLoadReadmeExample(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scenario.yaml")
	if err := os.WriteFile(path, []byte(readmeExample), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := scenario.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if s.Name != "scaling-agent-respects-quota-agent" {
		t.Errorf("name = %q", s.Name)
	}
	if len(s.Agents) != 2 {
		t.Errorf("agents = %v", s.Agents)
	}
	if len(s.Setup.Manifests) != 2 {
		t.Errorf("manifests = %v", s.Setup.Manifests)
	}
	if s.Trigger == nil || s.Trigger.Patch == nil {
		t.Fatal("expected trigger.patch")
	}
	if s.Trigger.Patch.Spec["replicas"] != 10 {
		t.Errorf("patch replicas = %v", s.Trigger.Patch.Spec["replicas"])
	}
	if s.Expect.Timeout != 120*time.Second {
		t.Errorf("timeout = %v", s.Expect.Timeout)
	}
	if len(s.Expect.Assertions) != 1 {
		t.Fatalf("assertions = %d", len(s.Expect.Assertions))
	}
	a := s.Expect.Assertions[0]
	if a.Resource.Name != "target" || len(a.Conditions) != 2 {
		t.Errorf("assertion = %+v", a)
	}

	paths, err := s.ResolveManifestPaths(scenario.BaseDir(path))
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(paths[0]) {
		t.Errorf("expected absolute manifest path, got %v", paths)
	}
}

func TestValidateRejectsMissingTimeout(t *testing.T) {
	s := &scenario.Scenario{
		Name:   "x",
		Agents: []string{"a"},
		Setup:  scenario.Setup{Manifests: []string{"m.yaml"}},
		Expect: scenario.Expect{
			Assertions: []scenario.ResourceAssertion{{
				Resource: scenario.ResourceRef{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "p",
					Namespace:  "ns",
				},
				Conditions: []scenario.PathCondition{{Path: ".metadata.name", Value: "p"}},
			}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected validation error for missing timeout")
	}
}

func TestExpectExplicitAssertionsKey(t *testing.T) {
	const yamlDoc = `name: explicit-assertions
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
  timeout: 30s
`
	dir := t.TempDir()
	path := filepath.Join(dir, "scenario.yaml")
	if err := os.WriteFile(path, []byte(yamlDoc), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := scenario.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Expect.Timeout != 30*time.Second || len(s.Expect.Assertions) != 1 {
		t.Fatalf("expect = %+v", s.Expect)
	}
}
