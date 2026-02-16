package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads a single scenario YAML file and returns the parsed Scenario.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading scenario %s: %w", path, err)
	}

	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing scenario %s: %w", path, err)
	}

	if s.Name == "" {
		return nil, fmt.Errorf("scenario %s: name is required", path)
	}
	if len(s.Agents) == 0 {
		return nil, fmt.Errorf("scenario %s: at least one agent is required", path)
	}

	return &s, nil
}

// LoadDir reads all .yaml and .yml files in a directory and returns the parsed scenarios.
func LoadDir(dir string) ([]*Scenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading scenario directory %s: %w", dir, err)
	}

	var scenarios []*Scenario
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		s, err := Load(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}

	return scenarios, nil
}
