package scenario_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitea.gitea/mirror/kube-agents-test/pkg/scenario"
	"gopkg.in/yaml.v3"
)

func writeScenarioFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func validScenarioBody(name string) string {
	return `name: ` + name + `
agents:
  - agent-a
setup:
  manifests:
    - fixtures/base.yaml
expect:
  timeout: 30s
  assertions:
    - resource:
        apiVersion: v1
        kind: Pod
        name: target
      conditions:
        - path: .status.phase
          value: Running
`
}

func TestLoadRejectsMalformedExpect(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name          string
		suffix        string
		wantInvalid   bool
	}{
		{
			name: "canonical_empty_assertions",
			suffix: `expect:
  timeout: 10s
  assertions: []
`,
		},
		{
			name: "canonical_missing_timeout",
			suffix: `expect:
  assertions:
    - resource: {apiVersion: v1, kind: Pod, name: p}
      conditions: [{path: .x, value: 1}]
`,
		},
		{
			name: "negative_timeout",
			suffix: `expect:
  timeout: -5s
  assertions:
    - resource: {apiVersion: v1, kind: Pod, name: p}
      conditions: [{path: .x, value: 1}]
`,
		},
		{
			name:        "null_expect",
			suffix:      `expect: null`,
			wantInvalid: true,
		},
		{
			name:   "scalar_expect",
			suffix: `expect: 120s`,
		},
		{
			name: "numeric_timeout_without_unit",
			suffix: `expect:
  timeout: 120
  assertions:
    - resource: {apiVersion: v1, kind: Pod, name: p}
      conditions: [{path: .x, value: 1}]
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			content := `name: malformed-expect
agents:
  - agent-a
setup:
  manifests:
    - fixtures/base.yaml
` + tc.suffix
			path := writeScenarioFile(t, dir, tc.name+".yaml", content)
			_, err := scenario.Load(path)
			if err == nil {
				t.Fatal("expected load/validation error")
			}
			if tc.wantInvalid && !errors.Is(err, scenario.ErrInvalidScenario) {
				t.Fatalf("errors.Is(ErrInvalidScenario)=false, got: %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidTriggers(t *testing.T) {
	dir := t.TempDir()
	base := validScenarioBody("trigger-tests")

	tests := []struct {
		name    string
		trigger string
	}{
		{name: "empty_trigger_object", trigger: "trigger: {}\n"},
		{
			name: "dual_patch_and_agent_restart",
			trigger: `trigger:
  patch: {apiVersion: v1, kind: Pod, name: p}
  agentRestart: {agentId: a}
`,
		},
		{
			name: "patch_missing_apiVersion",
			trigger: `trigger:
  patch: {kind: Pod, name: p}
`,
		},
		{
			name: "agent_restart_wrong_yaml_key",
			trigger: `trigger:
  agentRestart:
    agentID: leader
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := base + tc.trigger
			path := writeScenarioFile(t, dir, tc.name+".yaml", body)
			_, err := scenario.Load(path)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestUnmarshalDualTriggerRequiresValidate(t *testing.T) {
	const raw = `name: dual-trigger
agents: [a]
setup: {manifests: [m.yaml]}
trigger:
  patch: {apiVersion: v1, kind: Pod, name: p}
  agentRestart: {agentId: a}
expect:
  timeout: 1s
  assertions:
    - resource: {apiVersion: v1, kind: Pod, name: p}
      conditions: [{path: .x, value: 1}]
`
	var s scenario.Scenario
	if err := yaml.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Trigger == nil || s.Trigger.Patch == nil || s.Trigger.AgentRestart == nil {
		t.Fatal("yaml.Unmarshal should populate both trigger variants before validation")
	}
	err := s.Validate()
	if err == nil {
		t.Fatal("Validate must reject multiple trigger variants")
	}
	if !errors.Is(err, scenario.ErrInvalidScenario) {
		t.Fatalf("got %v", err)
	}
}

func TestLegacyExpectMappingAndDocSequence(t *testing.T) {
	t.Run("legacy_mapping", func(t *testing.T) {
		const y = `name: legacy-map
agents: [a]
setup: {manifests: [m.yaml]}
expect:
  resource: {apiVersion: v1, kind: Pod, name: p, namespace: ns}
  conditions:
    - path: .status.phase
      value: Running
  timeout: 45s
`
		var s scenario.Scenario
		if err := yaml.Unmarshal([]byte(y), &s); err != nil {
			t.Fatal(err)
		}
		if err := s.Validate(); err != nil {
			t.Fatal(err)
		}
		if s.Expect.Timeout != 45*time.Second {
			t.Errorf("timeout: got %v", s.Expect.Timeout)
		}
	})

	t.Run("doc_sequence", func(t *testing.T) {
		const y = `name: doc-seq
agents: [a]
setup: {manifests: [m.yaml]}
expect:
  - resource:
      apiVersion: apps/v1
      kind: Deployment
      name: target
      namespace: test
    conditions:
      - path: .spec.replicas
        value: 5
  - timeout: 120s
`
		var s scenario.Scenario
		if err := yaml.Unmarshal([]byte(y), &s); err != nil {
			t.Fatal(err)
		}
		if err := s.Validate(); err != nil {
			t.Fatal(err)
		}
		if s.Expect.Timeout != 120*time.Second {
			t.Errorf("timeout: got %v", s.Expect.Timeout)
		}
	})
}

func TestLoadDirNonRecursiveAndEmpty(t *testing.T) {
	dir := t.TempDir()

	scenarios, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatalf("empty dir: %v", err)
	}
	if len(scenarios) != 0 {
		t.Fatalf("empty dir: got %d scenarios", len(scenarios))
	}

	sub := filepath.Join(dir, "nested")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	nestedYAML := validScenarioBody("nested-only")
	if err := os.WriteFile(filepath.Join(sub, "hidden.yaml"), []byte(nestedYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	writeScenarioFile(t, dir, "top.yaml", validScenarioBody("top-level"))

	scenarios, err = scenario.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(scenarios) != 1 || scenarios[0].Name != "top-level" {
		t.Fatalf("LoadDir should be non-recursive: got %#v", scenarios)
	}
}

func TestLoadDirReturnsDuplicateNamesWithoutError(t *testing.T) {
	dir := t.TempDir()
	body := validScenarioBody("same-name")
	writeScenarioFile(t, dir, "one.yaml", body)
	writeScenarioFile(t, dir, "two.yml", body)

	scenarios, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(scenarios) != 2 {
		t.Fatalf("got %d scenarios", len(scenarios))
	}
	if scenarios[0].Name != scenarios[1].Name {
		t.Fatal("expected duplicate names in loaded suite")
	}
}
