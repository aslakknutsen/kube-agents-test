package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitea.gitea/mirror/kube-agents-test/pkg/fault"
	"gopkg.in/yaml.v3"
)

// Load reads and validates a scenario from a YAML file.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario %s: %w", path, err)
	}
	s, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse scenario %s: %w", path, err)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// LoadDir loads all .yaml and .yml scenario files in a directory (non-recursive).
func LoadDir(dir string) ([]*Scenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read scenario dir %s: %w", dir, err)
	}
	var out []*Scenario
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		s, err := Load(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func parse(data []byte) (*Scenario, error) {
	var raw scenarioYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw.toScenario()
}

type scenarioYAML struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Agents      []string       `yaml:"agents"`
	Setup       Setup          `yaml:"setup"`
	Trigger     *fault.Trigger `yaml:"trigger,omitempty"`
	Expect      expectYAML     `yaml:"expect"`
}

// expectYAML accepts canonical {timeout, assertions} or legacy sequence + timeout.
type expectYAML struct {
	Canonical *canonicalExpect `yaml:",inline"`
}

type canonicalExpect struct {
	Timeout    timeDuration `yaml:"timeout"`
	Assertions []Assertion  `yaml:"assertions"`
}

// UnmarshalYAML supports canonical and legacy expect shapes from docs.
func (e *expectYAML) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return fmt.Errorf("expect: empty node")
	}
	switch node.Kind {
	case yaml.MappingNode:
		var canon canonicalExpect
		if err := node.Decode(&canon); err != nil {
			return err
		}
		if len(canon.Assertions) > 0 || isCanonicalExpectMapping(node) {
			e.Canonical = &canon
			return nil
		}
		// Legacy mapping: top-level resource/conditions plus timeout sibling.
		var timeout timeDuration
		var assertions []Assertion
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			val := node.Content[i+1]
			switch key {
			case "timeout":
				if err := val.Decode(&timeout); err != nil {
					return err
				}
			case "assertions":
				if err := val.Decode(&assertions); err != nil {
					return err
				}
			case "resource":
				var a Assertion
				if err := val.Decode(&a.Resource); err != nil {
					return err
				}
				if i+2 < len(node.Content) && node.Content[i+2].Value == "conditions" {
					if err := node.Content[i+3].Decode(&a.Conditions); err != nil {
						return err
					}
					i += 2
				}
				assertions = append(assertions, a)
			case "conditions":
				if len(assertions) == 0 {
					return fmt.Errorf("expect: conditions without resource")
				}
				if err := val.Decode(&assertions[len(assertions)-1].Conditions); err != nil {
					return err
				}
			}
		}
		if len(assertions) > 0 {
			e.Canonical = &canonicalExpect{Timeout: timeout, Assertions: assertions}
			return nil
		}
		return fmt.Errorf("expect: unrecognized mapping shape")
	case yaml.SequenceNode:
		var timeout timeDuration
		var assertions []Assertion
		for _, child := range node.Content {
			if child.Kind == yaml.MappingNode && len(child.Content) == 2 &&
				child.Content[0].Value == "timeout" {
				if err := child.Content[1].Decode(&timeout); err != nil {
					return err
				}
				continue
			}
			var a Assertion
			if err := child.Decode(&a); err != nil {
				return err
			}
			if a.Resource.APIVersion != "" || len(a.Conditions) > 0 {
				assertions = append(assertions, a)
			}
		}
		if len(assertions) > 0 {
			e.Canonical = &canonicalExpect{Timeout: timeout, Assertions: assertions}
			return nil
		}
		return fmt.Errorf("expect: legacy sequence requires timeout and assertions")
	default:
		return fmt.Errorf("expect: expected mapping or sequence")
	}
}

func (raw *scenarioYAML) toScenario() (*Scenario, error) {
	exp, err := raw.Expect.toExpect()
	if err != nil {
		return nil, err
	}
	return &Scenario{
		Name:        raw.Name,
		Description: raw.Description,
		Agents:      raw.Agents,
		Setup:       raw.Setup,
		Trigger:     raw.Trigger,
		Expect:      exp,
	}, nil
}

// isCanonicalExpectMapping reports whether the YAML mapping uses only canonical
// expect keys (timeout, assertions), including empty assertions: [].
func isCanonicalExpectMapping(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.MappingNode || len(node.Content) == 0 {
		return false
	}
	for i := 0; i < len(node.Content); i += 2 {
		switch node.Content[i].Value {
		case "timeout", "assertions":
		default:
			return false
		}
	}
	return true
}

func (e *expectYAML) toExpect() (Expect, error) {
	if e.Canonical != nil {
		return Expect{
			Timeout:    e.Canonical.Timeout.Duration,
			Assertions: e.Canonical.Assertions,
		}, nil
	}
	return Expect{}, fmt.Errorf("%w: expect section is empty or invalid", ErrInvalidScenario)
}
