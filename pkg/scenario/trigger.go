package scenario

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Trigger describes an optional perturbation applied after setup.
type Trigger struct {
	Patch        *PatchTrigger        `yaml:"patch,omitempty"`
	AgentRestart *AgentRestartTrigger `yaml:"agentRestart,omitempty"`
	Fault        *FaultTrigger        `yaml:"fault,omitempty"`
}

// Validate ensures at most one trigger arm is set in v1.
func (t *Trigger) Validate() error {
	if t == nil {
		return nil
	}
	set := 0
	if t.Patch != nil {
		set++
	}
	if t.AgentRestart != nil {
		set++
	}
	if t.Fault != nil {
		set++
	}
	if set > 1 {
		return fmt.Errorf("trigger: at most one of patch, agentRestart, or fault may be set")
	}
	if t.AgentRestart != nil {
		if err := t.AgentRestart.Validate(); err != nil {
			return err
		}
	}
	if t.Fault != nil {
		if err := t.Fault.Validate(); err != nil {
			return err
		}
	}
	if t.Patch != nil {
		if err := t.Patch.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// PatchTrigger applies a resource patch via the Kubernetes API.
type PatchTrigger struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
	Raw        map[string]any
}

// UnmarshalYAML parses identity fields and stores remaining keys in Raw.
func (p *PatchTrigger) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("patch trigger: expected mapping, got %v", value.Kind)
	}
	type patchFields struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Name       string `yaml:"name"`
		Namespace  string `yaml:"namespace"`
	}
	var fields patchFields
	if err := value.Decode(&fields); err != nil {
		return err
	}
	p.APIVersion = fields.APIVersion
	p.Kind = fields.Kind
	p.Name = fields.Name
	p.Namespace = fields.Namespace

	raw := make(map[string]any)
	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valNode := value.Content[i+1]
		key := keyNode.Value
		switch key {
		case "apiVersion", "kind", "name", "namespace":
			continue
		default:
			var v any
			if err := valNode.Decode(&v); err != nil {
				return fmt.Errorf("patch trigger field %q: %w", key, err)
			}
			raw[key] = v
		}
	}
	p.Raw = raw
	return nil
}

// Validate checks required identity fields.
func (p *PatchTrigger) Validate() error {
	if p == nil {
		return nil
	}
	if p.APIVersion == "" {
		return fmt.Errorf("patch trigger: apiVersion is required")
	}
	if p.Kind == "" {
		return fmt.Errorf("patch trigger: kind is required")
	}
	if p.Name == "" {
		return fmt.Errorf("patch trigger: name is required")
	}
	return nil
}

// AgentRestartTrigger restarts an agent mid-scenario.
type AgentRestartTrigger struct {
	Agent string        `yaml:"agent"`
	Delay time.Duration `yaml:"delay,omitempty"`
}

// Validate checks required fields.
func (a *AgentRestartTrigger) Validate() error {
	if a == nil {
		return nil
	}
	if a.Agent == "" {
		return fmt.Errorf("agentRestart trigger: agent is required")
	}
	return nil
}

// FaultTrigger invokes a fault injection hook with extensible parameters.
type FaultTrigger struct {
	Type   string    `yaml:"type"`
	Params yaml.Node `yaml:"params,omitempty"`
}

// Validate checks the fault type is non-empty.
func (f *FaultTrigger) Validate() error {
	if f == nil {
		return nil
	}
	if f.Type == "" {
		return fmt.Errorf("fault trigger: type is required")
	}
	return nil
}
