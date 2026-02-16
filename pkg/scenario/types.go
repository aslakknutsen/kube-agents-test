package scenario

import "time"

// Scenario is the top-level structure parsed from a scenario YAML file.
type Scenario struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Agents      []string `yaml:"agents"`
	Setup       Setup    `yaml:"setup"`
	Trigger     *Trigger `yaml:"trigger,omitempty"`
	Expect      Expect   `yaml:"expect"`
}

// Setup defines the initial cluster state for a scenario.
type Setup struct {
	// Manifests is a list of file paths to Kubernetes YAML manifests
	// to apply before the test begins.
	Manifests []string `yaml:"manifests"`
}

// Trigger defines an optional mutation that kicks off the test.
type Trigger struct {
	Patch *PatchTrigger `yaml:"patch,omitempty"`
}

// PatchTrigger describes a resource patch to apply as the trigger.
type PatchTrigger struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// Expect defines the expected final state and convergence timeout.
type Expect struct {
	Resources []ResourceExpectation `yaml:"resources"`
	Timeout   Duration              `yaml:"timeout"`
}

// ResourceExpectation describes a single resource and the conditions
// it must satisfy for the scenario to pass.
type ResourceExpectation struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Name       string      `yaml:"name"`
	Namespace  string      `yaml:"namespace"`
	Conditions []Condition `yaml:"conditions"`
}

// Condition is a single field assertion on a resource.
type Condition struct {
	Path  string      `yaml:"path"`
	Value interface{} `yaml:"value"`
}

// Duration wraps time.Duration for YAML unmarshalling from strings like "120s".
type Duration struct {
	time.Duration
}

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
