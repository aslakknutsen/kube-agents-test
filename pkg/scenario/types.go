// Package scenario defines test scenario types, YAML loading, and the scenario engine API.
//
// API v0 — unstable until first release.
package scenario

import (
	"time"

	"gopkg.in/yaml.v3"
)

// Scenario is a declarative end-to-end test (see docs/scenarios.md).
type Scenario struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description,omitempty"`
	Agents      []string    `yaml:"agents"`
	Setup       Setup       `yaml:"setup"`
	Trigger     *Trigger    `yaml:"trigger,omitempty"`
	Expect      Expectation `yaml:"expect"`
}

// Setup describes initial cluster state before triggers and assertions.
type Setup struct {
	Manifests []string `yaml:"manifests"`
}

// Trigger perturbs the system under test. v1 supports patch; other kinds use Type + Raw.
type Trigger struct {
	Patch *ResourcePatch `yaml:"patch,omitempty"`

	// Type and Raw reserve agent restart, fault injection, and future trigger kinds.
	Type string     `yaml:"type,omitempty"`
	Raw  *yaml.Node `yaml:",inline,omitempty"`
}

// ResourcePatch applies a partial resource update (trigger.patch in scenario YAML).
type ResourcePatch struct {
	ObjectRef `yaml:",inline"`
	Body      map[string]any `yaml:",inline"`
}

// Expectation holds resource assertions and the convergence timeout.
type Expectation struct {
	Resources []ResourceExpect `yaml:"resources,omitempty"`
	Timeout   time.Duration      `yaml:"timeout"`
}

// ResourceExpect asserts JSONPath conditions on one Kubernetes resource.
type ResourceExpect struct {
	Resource   ObjectRef   `yaml:"resource"`
	Conditions []Condition `yaml:"conditions"`
}

// Condition is one JSONPath-based state assertion.
type Condition struct {
	Path  string `yaml:"path"`
	Value any    `yaml:"value"`
}

// ObjectRef identifies a Kubernetes API object.
type ObjectRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace,omitempty"`
}
