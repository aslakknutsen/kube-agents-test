package scenario

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML supports documented inline-list format and explicit resources+timeout.
//
// Documented format (docs/scenarios.md):
//
//	expect:
//	  - resource: ...
//	    conditions: ...
//	  timeout: 120s
func (e *Expectation) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}

	switch value.Kind {
	case yaml.SequenceNode:
		var resources []ResourceExpect
		if err := value.Decode(&resources); err != nil {
			return err
		}
		e.Resources = resources
		return nil
	case yaml.MappingNode:
		var explicit struct {
			Resources []ResourceExpect `yaml:"resources"`
			Timeout   time.Duration    `yaml:"timeout"`
		}
		if err := value.Decode(&explicit); err != nil {
			return err
		}
		if len(explicit.Resources) > 0 || explicit.Timeout > 0 {
			e.Resources = explicit.Resources
			e.Timeout = explicit.Timeout
			return nil
		}
		return e.decodeMixedMapping(value)
	default:
		return fmt.Errorf("expect: unsupported YAML kind %v", value.Kind)
	}
}

func (e *Expectation) decodeMixedMapping(node *yaml.Node) error {
	var timeout time.Duration
	var resources []ResourceExpect

	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "timeout":
			if err := val.Decode(&timeout); err != nil {
				return fmt.Errorf("expect.timeout: %w", err)
			}
		default:
			if val.Kind == yaml.SequenceNode {
				if err := val.Decode(&resources); err != nil {
					return fmt.Errorf("expect resources: %w", err)
				}
			} else if val.Kind == yaml.MappingNode {
				var re ResourceExpect
				if err := val.Decode(&re); err != nil {
					return fmt.Errorf("expect resource: %w", err)
				}
				resources = append(resources, re)
			}
		}
	}

	e.Resources = resources
	e.Timeout = timeout
	return nil
}
