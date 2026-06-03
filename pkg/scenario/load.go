package scenario

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a scenario YAML file from path.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario %s: %w", path, err)
	}
	data = rewriteExpectLegacy(data)
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scenario %s: %w", path, err)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// BaseDir returns the directory containing the scenario file, used to resolve
// relative manifest paths.
func BaseDir(scenarioPath string) string {
	return filepath.Dir(scenarioPath)
}

// ResolveManifestPaths returns absolute paths for setup manifests relative to
// the scenario file directory.
func (s *Scenario) ResolveManifestPaths(baseDir string) ([]string, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: scenario is nil", ErrInvalidScenario)
	}
	out := make([]string, 0, len(s.Setup.Manifests))
	for _, m := range s.Setup.Manifests {
		if filepath.IsAbs(m) {
			out = append(out, m)
			continue
		}
		out = append(out, filepath.Join(baseDir, m))
	}
	return out, nil
}
