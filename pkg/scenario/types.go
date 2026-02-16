// Package scenario defines the data types for test scenario definitions.
package scenario

import "time"

// Scenario is a declarative description of a multi-agent test case.
// It maps directly to the YAML scenario files.
type Scenario struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Agents      []string      `yaml:"agents"`
	Setup       Setup         `yaml:"setup"`
	Trigger     *Trigger      `yaml:"trigger,omitempty"`
	Expect      []Expectation `yaml:"expect"`
	Timeout     Duration      `yaml:"timeout"`
}

// Setup describes the initial cluster state for a scenario.
type Setup struct {
	// Manifests are paths to YAML files to apply before the test begins.
	Manifests []string `yaml:"manifests"`
}

// Trigger describes a mutation that kicks off the test after setup.
type Trigger struct {
	// Patch applies a strategic merge patch to an existing resource.
	Patch *ResourcePatch `yaml:"patch,omitempty"`
}

// ResourcePatch identifies a resource and the fields to patch.
type ResourcePatch struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

// Expectation describes the expected state of a single resource after convergence.
type Expectation struct {
	Resource   ResourceRef `yaml:"resource"`
	Conditions []Condition `yaml:"conditions"`
}

// ResourceRef identifies a Kubernetes resource.
type ResourceRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
}

// Condition is a single field-level assertion on a resource.
type Condition struct {
	// Path is a dot-separated JSONPath-like expression (e.g. ".spec.replicas").
	Path  string      `yaml:"path"`
	Value interface{} `yaml:"value"`
}

// Duration wraps time.Duration with YAML unmarshalling from strings like "120s".
type Duration struct {
	time.Duration
}

// UnmarshalYAML parses duration strings like "30s", "2m", "1h".
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}
