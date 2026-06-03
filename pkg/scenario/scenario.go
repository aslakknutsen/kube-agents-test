// Package scenario defines YAML domain types for test scenarios.
// See docs/scenarios.md for the scenario file format.
package scenario

import (
	"errors"
	"fmt"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/fault"
	"gopkg.in/yaml.v3"
)

// ErrInvalidScenario indicates the scenario failed structural validation.
var ErrInvalidScenario = errors.New("invalid scenario")

// Scenario is the unit of work — one YAML file, one behavioral expectation.
type Scenario struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Agents      []string `yaml:"agents"`
	Setup       Setup    `yaml:"setup"`
	Trigger     *Trigger `yaml:"trigger,omitempty"`
	Expect      Expect   `yaml:"expect"`
}

// Setup describes initial cluster state.
type Setup struct {
	Manifests []string `yaml:"manifests"`
}

// Trigger is a discriminated union of perturbation mechanisms.
type Trigger struct {
	Patch       *ResourcePatch `yaml:"patch,omitempty"`
	AgentAction *AgentAction   `yaml:"agentAction,omitempty"`
	Fault       *fault.Spec    `yaml:"fault,omitempty"`
}

// AgentAction describes agent lifecycle triggers (restart/stop); schema may evolve.
type AgentAction struct {
	Agent  string `yaml:"agent"`
	Action string `yaml:"action"` // e.g. restart, stop
}

// ResourcePatch applies a partial resource update.
type ResourcePatch struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

// Expect holds assertions and a convergence timeout.
type Expect struct {
	Assertions []ResourceAssertion `yaml:"assertions"`
	Timeout    time.Duration       `yaml:"timeout"`
}

// ResourceAssertion checks resource fields via JSONPath conditions.
type ResourceAssertion struct {
	Resource   ResourceRef     `yaml:"resource"`
	Conditions []PathCondition `yaml:"conditions"`
}

// ResourceRef identifies a Kubernetes object.
type ResourceRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
}

// PathCondition asserts a JSONPath evaluates to an expected value.
type PathCondition struct {
	Path  string      `yaml:"path"`
	Value interface{} `yaml:"value"`
}

// UnmarshalYAML decodes expect with explicit assertions and timeout fields.
func (e *Expect) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: expect must be a mapping", ErrInvalidScenario)
	}

	var timeout time.Duration
	var assertions []ResourceAssertion

	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i].Value
		val := value.Content[i+1]
		switch key {
		case "timeout":
			var raw string
			if err := val.Decode(&raw); err != nil {
				return fmt.Errorf("%w: decode expect.timeout: %v", ErrInvalidScenario, err)
			}
			d, err := time.ParseDuration(raw)
			if err != nil {
				return fmt.Errorf("%w: parse expect.timeout: %v", ErrInvalidScenario, err)
			}
			timeout = d
		case "assertions":
			if err := val.Decode(&assertions); err != nil {
				return fmt.Errorf("%w: decode expect.assertions: %v", ErrInvalidScenario, err)
			}
		default:
			return fmt.Errorf("%w: unknown expect field %q", ErrInvalidScenario, key)
		}
	}

	e.Timeout = timeout
	e.Assertions = assertions
	return nil
}

// Validate checks required fields and basic consistency.
func (s *Scenario) Validate() error {
	if s == nil {
		return fmt.Errorf("%w: scenario is nil", ErrInvalidScenario)
	}
	if s.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidScenario)
	}
	if len(s.Agents) == 0 {
		return fmt.Errorf("%w: agents must not be empty", ErrInvalidScenario)
	}
	if len(s.Setup.Manifests) == 0 {
		return fmt.Errorf("%w: setup.manifests must not be empty", ErrInvalidScenario)
	}
	if s.Expect.Timeout <= 0 {
		return fmt.Errorf("%w: expect.timeout must be positive", ErrInvalidScenario)
	}
	if len(s.Expect.Assertions) == 0 {
		return fmt.Errorf("%w: expect must contain at least one assertion", ErrInvalidScenario)
	}
	for i, a := range s.Expect.Assertions {
		if err := validateResourceRef(a.Resource, fmt.Sprintf("expect.assertions[%d].resource", i)); err != nil {
			return err
		}
		if len(a.Conditions) == 0 {
			return fmt.Errorf("%w: expect.assertions[%d] must have conditions", ErrInvalidScenario, i)
		}
		for j, c := range a.Conditions {
			if c.Path == "" {
				return fmt.Errorf("%w: expect.assertions[%d].conditions[%d].path is required", ErrInvalidScenario, i, j)
			}
		}
	}
	if s.Trigger != nil && s.Trigger.Patch != nil {
		if err := validateResourceRef(resourceRefFromPatch(*s.Trigger.Patch), "trigger.patch"); err != nil {
			return err
		}
	}
	return nil
}

func validateResourceRef(r ResourceRef, field string) error {
	if r.APIVersion == "" || r.Kind == "" || r.Name == "" || r.Namespace == "" {
		return fmt.Errorf("%w: %s requires apiVersion, kind, name, and namespace", ErrInvalidScenario, field)
	}
	return nil
}

func resourceRefFromPatch(p ResourcePatch) ResourceRef {
	return ResourceRef{
		APIVersion: p.APIVersion,
		Kind:       p.Kind,
		Name:       p.Name,
		Namespace:  p.Namespace,
	}
}
