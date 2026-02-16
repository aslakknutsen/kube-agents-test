package scenario

import "time"

// Scenario is the top-level structure for a test scenario YAML file.
type Scenario struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Agents      []string `yaml:"agents"`
	Setup       Setup    `yaml:"setup"`
	Trigger     *Trigger `yaml:"trigger,omitempty"`
	Expect      Expect   `yaml:"expect"`
}

// Setup defines the initial cluster state.
type Setup struct {
	Manifests []string `yaml:"manifests"`
}

// Trigger defines an optional mutation that kicks off the scenario.
type Trigger struct {
	Patch *ResourcePatch `yaml:"patch,omitempty"`
}

// ResourcePatch describes a patch to apply to a Kubernetes resource.
type ResourcePatch struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// Expect defines the expected final state and the convergence timeout.
type Expect struct {
	Resources []ResourceExpectation `yaml:"resources"`
	Timeout   Duration              `yaml:"timeout"`
}

// ResourceExpectation is a single resource to check.
type ResourceExpectation struct {
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

// Condition checks a specific field path against an expected value.
type Condition struct {
	Path  string      `yaml:"path"`
	Value interface{} `yaml:"value"`
}

// Duration is a wrapper around time.Duration that supports YAML unmarshaling
// from strings like "120s".
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}
