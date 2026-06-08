package scenario

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadedScenario pairs a parsed scenario with its source file path.
type LoadedScenario struct {
	Path     string
	Scenario *Scenario
}

// Load reads and validates a scenario YAML file.
func Load(path string) (*Scenario, error) {
	loaded, err := loadFile(path)
	if err != nil {
		return nil, err
	}
	return loaded.Scenario, nil
}

func loadFile(path string) (LoadedScenario, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return LoadedScenario{}, fmt.Errorf("scenario path: %w", err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return LoadedScenario{}, fmt.Errorf("read scenario %q: %w", path, err)
	}
	data = normalizeScenarioYAML(data)
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return LoadedScenario{}, fmt.Errorf("parse scenario %q: %w", path, err)
	}
	if err := s.Validate(); err != nil {
		return LoadedScenario{}, err
	}
	scenarioDir := filepath.Dir(absPath)
	if err := ValidateManifestPaths(scenarioDir, s.Setup.Manifests); err != nil {
		return LoadedScenario{}, fmt.Errorf("scenario %q: %w", s.Name, err)
	}
	return LoadedScenario{Path: absPath, Scenario: &s}, nil
}

// LoadDir loads all scenario YAML files under dir recursively.
func LoadDir(dir string) ([]LoadedScenario, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("scenario directory: %w", err)
	}

	var paths []string
	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk scenario directory %q: %w", dir, err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no scenario files found in %q", dir)
	}
	sort.Strings(paths)

	loaded := make([]LoadedScenario, 0, len(paths))
	for _, path := range paths {
		item, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, item)
	}
	return loaded, nil
}

// LoadPaths loads scenarios from files and directories in stable sorted order.
func LoadPaths(paths []string) ([]LoadedScenario, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("no scenario paths provided")
	}

	var all []LoadedScenario
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("scenario path %q: %w", p, err)
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", p, err)
		}
		if info.IsDir() {
			scenarios, err := LoadDir(absPath)
			if err != nil {
				return nil, err
			}
			all = append(all, scenarios...)
			continue
		}
		item, err := loadFile(absPath)
		if err != nil {
			return nil, err
		}
		all = append(all, item)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].Scenario.Name != all[j].Scenario.Name {
			return all[i].Scenario.Name < all[j].Scenario.Name
		}
		return all[i].Path < all[j].Path
	})

	return all, nil
}
