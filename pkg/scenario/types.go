package scenario

import (
	"fmt"
	"strings"
	"time"

	"gitea.gitea/mirror/kube-agents-test/pkg/fault"
	"gopkg.in/yaml.v3"
)

// Document wraps a scenario file with optional API metadata.
type Document struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string `yaml:"kind,omitempty"`
}

// Scenario is a declarative end-to-end test definition.
type Scenario struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Agents      []string       `yaml:"agents"`
	Setup       Setup          `yaml:"setup"`
	Trigger     *fault.Trigger `yaml:"trigger,omitempty"`
	Expect      Expect         `yaml:"expect"`
}

// Setup describes initial cluster state.
type Setup struct {
	Manifests []string `yaml:"manifests"`
}

// Expect holds convergence timeout and resource assertions.
type Expect struct {
	Timeout    time.Duration `yaml:"timeout"`
	Assertions []Assertion   `yaml:"assertions"`
}

// UnmarshalYAML supports canonical and legacy expect shapes (see load.go).
func (e *Expect) UnmarshalYAML(node *yaml.Node) error {
	var raw expectYAML
	if err := raw.UnmarshalYAML(node); err != nil {
		return err
	}
	parsed, err := raw.toExpect()
	if err != nil {
		return err
	}
	*e = parsed
	return nil
}

// Assertion groups conditions that must all pass on one resource.
type Assertion struct {
	Resource   ResourceRef `yaml:"resource"`
	Conditions []Condition `yaml:"conditions"`
}

// ResourceRef identifies a Kubernetes object.
type ResourceRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
}

// Condition is a JSONPath equality check against a resource.
type Condition struct {
	Path  string      `yaml:"path"`
	Value interface{} `yaml:"value"`
}

// Validate checks required fields and nested structures.
func (s *Scenario) Validate() error {
	if s == nil {
		return fmt.Errorf("%w: scenario is nil", ErrInvalidScenario)
	}
	if isBlank(s.Name) {
		return fmt.Errorf("%w: name is required", ErrInvalidScenario)
	}
	if len(s.Agents) == 0 {
		return fmt.Errorf("%w: at least one agent is required", ErrInvalidScenario)
	}
	for i, id := range s.Agents {
		if isBlank(id) {
			return fmt.Errorf("%w: agents[%d] is empty", ErrInvalidScenario, i)
		}
	}
	if len(s.Setup.Manifests) == 0 {
		return fmt.Errorf("%w: setup.manifests is required", ErrInvalidScenario)
	}
	for i, p := range s.Setup.Manifests {
		if isBlank(p) {
			return fmt.Errorf("%w: setup.manifests[%d] is empty", ErrInvalidScenario, i)
		}
	}
	if s.Trigger != nil {
		if err := s.Trigger.Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidScenario, err)
		}
	}
	if s.Expect.Timeout <= 0 {
		return fmt.Errorf("%w: expect.timeout must be positive", ErrInvalidScenario)
	}
	if len(s.Expect.Assertions) == 0 {
		return fmt.Errorf("%w: expect.assertions is required", ErrInvalidScenario)
	}
	for i, a := range s.Expect.Assertions {
		if err := a.validate(); err != nil {
			return fmt.Errorf("%w: expect.assertions[%d]: %w", ErrInvalidScenario, i, err)
		}
	}
	return nil
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func (a *Assertion) validate() error {
	if isBlank(a.Resource.APIVersion) || isBlank(a.Resource.Kind) || isBlank(a.Resource.Name) {
		return fmt.Errorf("assertion resource requires apiVersion, kind, and name")
	}
	if len(a.Conditions) == 0 {
		return fmt.Errorf("assertion requires at least one condition")
	}
	for i, c := range a.Conditions {
		if isBlank(c.Path) {
			return fmt.Errorf("conditions[%d].path is required", i)
		}
	}
	return nil
}
