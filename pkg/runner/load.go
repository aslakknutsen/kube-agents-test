package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// LoadedScenario pairs a scenario with its source file path for manifest resolution.
type LoadedScenario struct {
	Path     string
	Scenario *scenario.Scenario
}

// LoadPath loads scenarios from a single YAML file or a directory (non-recursive).
func LoadPath(path string) ([]LoadedScenario, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("load scenarios: %w", err)
	}
	if info.IsDir() {
		return loadDir(path)
	}
	sc, err := scenario.LoadFile(path)
	if err != nil {
		return nil, err
	}
	return []LoadedScenario{{Path: path, Scenario: sc}}, nil
}

func loadDir(dir string) ([]LoadedScenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read scenario dir %s: %w", dir, err)
	}
	var out []LoadedScenario
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, name)
		sc, err := scenario.LoadFile(path)
		if err != nil {
			return nil, err
		}
		out = append(out, LoadedScenario{Path: path, Scenario: sc})
	}
	return out, nil
}

// BaseDirForScenario returns the directory used to resolve relative manifest paths.
func BaseDirForScenario(path string) string {
	return filepath.Dir(path)
}
