package scenario

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFile reads and parses one scenario YAML file.
func LoadFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario %s: %w", path, err)
	}
	var sc Scenario
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("parse scenario %s: %w", path, err)
	}
	if err := sc.Validate(); err != nil {
		return nil, fmt.Errorf("scenario %s: %w", path, err)
	}
	return &sc, nil
}

// LoadDir loads all *.yaml and *.yml files in dir (non-recursive).
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
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		sc, err := LoadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, nil
}
