package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/framework"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

var fw *framework.Framework

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Default to kind. Set KUBECONFIG env var to use an existing cluster.
	var provider cluster.Provider
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		provider = cluster.NewExistingClusterProvider(kc)
	} else {
		provider = cluster.NewKindProvider("kube-agents-test")
	}

	var err error
	fw, err = framework.New(ctx, framework.Config{
		ClusterProvider: provider,
	})
	if err != nil {
		panic("failed to create framework: " + err.Error())
	}

	code := m.Run()

	_ = fw.Teardown(ctx)
	os.Exit(code)
}

func TestScalingAgentRespectsQuotaAgent(t *testing.T) {
	fw.RunScenario(t, filepath.Join("..", "scenarios", "scaling-agent-respects-quota-agent.yaml"))
}

func TestScenarioYAMLParsing(t *testing.T) {
	scenarioDir := filepath.Join("..", "scenarios")
	entries, err := os.ReadDir(scenarioDir)
	if err != nil {
		t.Fatalf("reading scenarios dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(scenarioDir, entry.Name())
			s, err := scenario.ParseFile(path)
			if err != nil {
				t.Fatalf("failed to parse %s: %v", path, err)
			}
			if s.Name == "" {
				t.Error("scenario name is empty")
			}
			if len(s.Agents) == 0 {
				t.Error("scenario has no agents")
			}
			if len(s.Expect.Resources) == 0 {
				t.Error("scenario has no expectations")
			}
		})
	}
}

func TestScenarioParserRejectsEmpty(t *testing.T) {
	_, err := scenario.Parse([]byte(""))
	if err == nil {
		t.Error("expected error for empty YAML, got nil")
	}
}

func TestScenarioParserRejectsNoName(t *testing.T) {
	yaml := []byte(`
description: missing name field
agents:
  - some-agent
expect:
  timeout: 30s
`)
	_, err := scenario.Parse(yaml)
	if err == nil {
		t.Error("expected error for scenario without name, got nil")
	}
}
