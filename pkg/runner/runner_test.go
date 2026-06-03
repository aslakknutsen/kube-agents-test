package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	agentstub "github.com/kube-agents/kube-agents-test/internal/agent/stub"
	clusterstub "github.com/kube-agents/kube-agents-test/internal/cluster/stub"
	enginestub "github.com/kube-agents/kube-agents-test/internal/engine/stub"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
)

func TestRunPathSingleScenario(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scenario.yaml")
	if err := os.WriteFile(path, []byte(minimalScenario), 0o644); err != nil {
		t.Fatal(err)
	}

	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &clusterstub.Provider{},
		AgentManager:    &agentstub.Manager{},
		Engine:          &enginestub.Engine{},
	})

	report, err := r.RunPath(context.Background(), path)
	if err != nil {
		t.Fatalf("RunPath: %v", err)
	}
	if report.Summary.Total != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}

func TestNewRunnerImplementsInterface(t *testing.T) {
	var _ runner.Runner = runner.New(runner.Config{}, runner.Dependencies{})
}

const minimalScenario = `name: minimal
agents:
  - scaling-agent
setup:
  manifests:
    - fixtures/setup.yaml
expect:
  - resource:
      apiVersion: v1
      kind: ConfigMap
      name: cfg
      namespace: test
    conditions:
      - path: .metadata.name
        value: cfg
  timeout: 10s
`

func TestRunPathDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(minimalScenario), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runner.New(runner.Config{}, runner.Dependencies{
		ClusterProvider: &clusterstub.Provider{},
		AgentManager:    &agentstub.Manager{},
		Engine:          &enginestub.Engine{},
	})
	report, err := r.RunPath(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.Total != 1 {
		t.Fatalf("summary = %+v", report.Summary)
	}
}
