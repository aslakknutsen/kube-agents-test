package scenario

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Expect holds resource assertions and a convergence timeout.
type Expect struct {
	Assertions []ResourceAssertion
	Timeout    time.Duration
}

// UnmarshalYAML accepts documented mixed shapes and a normalized mapping form.
func (e *Expect) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return fmt.Errorf("expect: nil node")
	}

	switch value.Kind {
	case yaml.MappingNode:
		return e.unmarshalMapping(value)
	case yaml.SequenceNode:
		return e.unmarshalSequence(value)
	default:
		return fmt.Errorf("expect: unsupported node kind %v", value.Kind)
	}
}

func (e *Expect) unmarshalMapping(node *yaml.Node) error {
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		val := node.Content[i+1]
		switch key {
		case "timeout":
			var d time.Duration
			if err := val.Decode(&d); err != nil {
				return fmt.Errorf("expect timeout: %w", err)
			}
			e.Timeout = d
		case "assertions":
			var assertions []ResourceAssertion
			if err := val.Decode(&assertions); err != nil {
				return fmt.Errorf("expect assertions: %w", err)
			}
			e.Assertions = append(e.Assertions, assertions...)
		default:
			var assertion ResourceAssertion
			if err := val.Decode(&assertion); err != nil {
				return fmt.Errorf("expect key %q: %w", key, err)
			}
			if assertion.Resource.APIVersion != "" || assertion.Resource.Kind != "" {
				e.Assertions = append(e.Assertions, assertion)
			}
		}
	}
	return nil
}

func (e *Expect) unmarshalSequence(node *yaml.Node) error {
	for _, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return fmt.Errorf("expect: sequence item must be a mapping")
		}
		var assertion ResourceAssertion
		if err := item.Decode(&assertion); err != nil {
			return err
		}
		if assertion.Resource.APIVersion == "" && assertion.Resource.Kind == "" {
			continue
		}
		e.Assertions = append(e.Assertions, assertion)
	}
	return nil
}

// ResourceAssertion targets a resource and checks JSONPath conditions.
type ResourceAssertion struct {
	Resource   ResourceSelector `yaml:"resource"`
	Conditions []Condition      `yaml:"conditions"`
}

// ResourceSelector identifies a Kubernetes object.
type ResourceSelector struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
}

// Condition is a JSONPath assertion on a resource field.
type Condition struct {
	Path  string    `yaml:"path"`
	Value yaml.Node `yaml:"value"`
}
