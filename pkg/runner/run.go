package runner

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func runScenarios(ctx context.Context, deps Dependencies, scenarios []*scenario.Scenario, baseDir string) (*RunReport, error) {
	bases := make([]string, len(scenarios))
	for i := range bases {
		bases[i] = baseDir
	}
	return runScenariosWithBases(ctx, deps, scenarios, bases)
}

func runScenariosWithBases(ctx context.Context, deps Dependencies, scenarios []*scenario.Scenario, baseDirs []string) (*RunReport, error) {
	if deps.ClusterProvider == nil || deps.AgentManager == nil || deps.Engine == nil {
		return nil, fmt.Errorf("runner dependencies must include ClusterProvider, AgentManager, and Engine")
	}

	report := &RunReport{
		Results: make([]*engine.Result, 0, len(scenarios)),
	}

	for i, sc := range scenarios {
		base := ""
		if i < len(baseDirs) {
			base = baseDirs[i]
		}

		if sc == nil {
			report.Results = append(report.Results, &engine.Result{
				Status: engine.StatusError,
				Err:    fmt.Errorf("scenario at index %d is nil", i),
			})
			report.Summary.Total++
			report.Summary.Errored++
			continue
		}

		result, err := runOne(ctx, deps, sc, base)
		if err != nil {
			return report, err
		}
		report.Results = append(report.Results, result)
		report.Summary.Total++
		switch result.Status {
		case engine.StatusPassed:
			report.Summary.Passed++
		case engine.StatusFailed:
			report.Summary.Failed++
		case engine.StatusError:
			report.Summary.Errored++
		}
	}

	return report, nil
}

func runOne(ctx context.Context, deps Dependencies, sc *scenario.Scenario, baseDir string) (*engine.Result, error) {
	handle, err := deps.ClusterProvider.Acquire(ctx)
	if err != nil {
		return &engine.Result{
			ScenarioName: sc.Name,
			Status:       engine.StatusError,
			Err:          err,
		}, nil
	}
	defer handle.Release(ctx)

	if err := deps.AgentManager.DeploySet(ctx, sc.Agents); err != nil {
		return &engine.Result{
			ScenarioName: sc.Name,
			Status:       engine.StatusError,
			Err:          err,
		}, nil
	}
	defer deps.AgentManager.Teardown(ctx)

	return deps.Engine.Execute(ctx, engine.ExecuteRequest{
		Scenario: sc,
		BaseDir:  baseDir,
		Cluster:  handle,
		Agents:   deps.AgentManager,
	})
}

func loadScenariosFromPath(path string) ([]*scenario.Scenario, []string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}

	if !info.IsDir() {
		s, err := scenario.Load(path)
		if err != nil {
			return nil, nil, err
		}
		return []*scenario.Scenario{s}, []string{scenario.BaseDir(path)}, nil
	}

	var scenarios []*scenario.Scenario
	var bases []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !isScenarioFile(p) {
			return nil
		}
		s, loadErr := scenario.Load(p)
		if loadErr != nil {
			return loadErr
		}
		scenarios = append(scenarios, s)
		bases = append(bases, scenario.BaseDir(p))
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(scenarios) == 0 {
		return nil, nil, fmt.Errorf("no scenario YAML files found under %s", path)
	}
	return scenarios, bases, nil
}

func isScenarioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
