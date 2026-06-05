package scenario_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func validScenario() *scenario.Scenario {
	return &scenario.Scenario{
		Name:   "valid-scenario",
		Agents: scenario.AgentSet{"agent-a"},
		Setup:  scenario.Setup{Manifests: []string{"fixtures/setup.yaml"}},
		Expect: scenario.Expect{
			Assertions: []scenario.ResourceAssertion{{
				Resource: scenario.ResourceSelector{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "p",
					Namespace:  "sandbox",
				},
				Conditions: []scenario.Condition{{Path: ".metadata.name"}},
			}},
			Timeout: time.Minute,
		},
	}
}

func TestValidateRejectsInvalidScenarioNames(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"../../artifacts/evil", "invalid characters"},
		{"has spaces", "invalid characters"},
		{"", "name is required"},
		{"foo/bar", "invalid characters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validScenario()
			s.Name = tc.name
			err := s.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %q, want substring %q", err, tc.want)
			}
		})
	}
}

func TestValidateRejectsEmptyAgentEntries(t *testing.T) {
	s := validScenario()
	s.Agents = scenario.AgentSet{"agent-a", "  "}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for whitespace-only agent id")
	}
}

func TestValidateWithRejectsOutOfSandboxNamespace(t *testing.T) {
	s := validScenario()
	s.Expect.Assertions[0].Resource.Namespace = "kube-system"

	err := s.ValidateWith(scenario.ValidationContext{
		SandboxNamespace:   "sandbox",
		AllowClusterScoped: false,
	})
	if err == nil {
		t.Fatal("expected sandbox namespace violation")
	}
	if !strings.Contains(err.Error(), "outside sandbox") {
		t.Fatalf("error = %q", err)
	}
}

func TestValidateWithRejectsClusterScopedWithoutFlag(t *testing.T) {
	s := validScenario()
	s.Expect.Assertions[0].Resource.Namespace = ""

	err := s.ValidateWith(scenario.ValidationContext{
		SandboxNamespace:   "sandbox",
		AllowClusterScoped: false,
	})
	if err == nil {
		t.Fatal("expected cluster-scoped resource rejection")
	}
}

func TestValidateWithRejectsFaultWithoutAllowFaults(t *testing.T) {
	s := validScenario()
	s.Trigger = &scenario.Trigger{
		Fault: &scenario.FaultTrigger{Type: "kill-agent"},
	}

	err := s.ValidateWith(scenario.ValidationContext{
		SandboxNamespace: "sandbox",
		AllowFaults:        false,
	})
	if err == nil {
		t.Fatal("expected fault trigger rejection when AllowFaults is false")
	}
}

func TestValidateManifestPathsRejectsAbsolutePath(t *testing.T) {
	err := scenario.ValidateManifestPaths("/tmp/scenarios", []string{"/etc/passwd"})
	if err == nil {
		t.Fatal("expected error for absolute manifest path")
	}
	if !strings.Contains(err.Error(), "relative") {
		t.Fatalf("error = %q", err)
	}
}

func TestValidateManifestPathsRejectsDotDotInMiddle(t *testing.T) {
	err := scenario.ValidateManifestPaths("/tmp/scenarios", []string{"fixtures/../../secret.yaml"})
	if err == nil {
		t.Fatal("expected error for embedded .. segment")
	}
}

func TestValidateRejectsZeroTimeout(t *testing.T) {
	s := validScenario()
	s.Expect.Timeout = 0
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestValidateRejectsNegativeTimeout(t *testing.T) {
	s := validScenario()
	s.Expect.Timeout = -time.Second
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

func TestValidateRejectsEmptyExpectAssertions(t *testing.T) {
	s := validScenario()
	s.Expect.Assertions = nil
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for empty assertions")
	}
}

func TestValidateRejectsEmptySetupManifests(t *testing.T) {
	s := validScenario()
	s.Setup.Manifests = nil
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for empty setup.manifests")
	}
}

func TestPatchTriggerValidateRequiresIdentityFields(t *testing.T) {
	p := &scenario.PatchTrigger{Kind: "Pod", Name: "p"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error when apiVersion missing")
	}
}
