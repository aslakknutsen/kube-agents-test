package scenario

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads a scenario YAML file and returns a Scenario.
func ParseFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading scenario file %s: %w", path, err)
	}
	return Parse(data)
}

// Parse unmarshals YAML bytes into a Scenario.
func Parse(data []byte) (*Scenario, error) {
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing scenario YAML: %w", err)
	}
	if s.Name == "" {
		return nil, fmt.Errorf("scenario must have a name")
	}
	return &s, nil
}
