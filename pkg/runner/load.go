package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// LoadPath loads scenarios from a single YAML file or a directory (non-recursive).
func LoadPath(path string) ([]*scenario.Scenario, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("load scenarios: %w", err)
	}
	if info.IsDir() {
		return scenario.LoadDir(path)
	}
	sc, err := scenario.LoadFile(path)
	if err != nil {
		return nil, err
	}
	return []*scenario.Scenario{sc}, nil
}

// BaseDirForScenario returns the directory used to resolve relative manifest paths.
func BaseDirForScenario(path string) string {
	return filepath.Dir(path)
}
