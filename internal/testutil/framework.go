// Package testutil provides helpers for writing kube-agents integration tests
// using Go's standard testing package.
package testutil

import (
	"context"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Framework ties together all the components needed to run scenario tests.
type Framework struct {
	ClusterProvider cluster.Provider
	AgentManager    agent.Manager
	Collector       diagnostics.Collector
	Engine          *engine.Engine

	kubeconfigPath string
}

// Options configures the test framework.
type Options struct {
	// ClusterProvider is the cluster lifecycle manager.
	// If nil, defaults to KindProvider.
	ClusterProvider cluster.Provider

	// AgentManager manages agent deployment.
	// If nil, a PodManager is created after the cluster is up.
	AgentManager agent.Manager

	// Collector gathers diagnostics on failure.
	// If nil, a ClusterCollector is created automatically.
	Collector diagnostics.Collector

	// AgentConfigs are used to create a PodManager when AgentManager is nil.
	AgentConfigs []agent.AgentConfig
}

// Setup initializes the framework: provisions a cluster and wires up components.
// Call this once in TestMain or at the start of a test suite.
func Setup(t *testing.T, opts Options) *Framework {
	t.Helper()

	ctx := context.Background()
	f := &Framework{}

	// Cluster provider.
	if opts.ClusterProvider != nil {
		f.ClusterProvider = opts.ClusterProvider
	} else {
		f.ClusterProvider = &cluster.KindProvider{}
	}

	kubeconfigPath, err := f.ClusterProvider.Create(ctx)
	if err != nil {
		t.Fatalf("creating cluster: %v", err)
	}
	f.kubeconfigPath = kubeconfigPath

	t.Cleanup(func() {
		if err := f.ClusterProvider.Destroy(context.Background()); err != nil {
			t.Errorf("destroying cluster: %v", err)
		}
	})

	// Agent manager.
	if opts.AgentManager != nil {
		f.AgentManager = opts.AgentManager
	} else {
		f.AgentManager = agent.NewPodManager(kubeconfigPath, opts.AgentConfigs)
	}

	// Diagnostics collector.
	if opts.Collector != nil {
		f.Collector = opts.Collector
	} else {
		f.Collector = diagnostics.NewClusterCollector(kubeconfigPath)
	}

	// Engine.
	f.Engine = &engine.Engine{
		KubeconfigPath: kubeconfigPath,
		AgentManager:   f.AgentManager,
		Collector:      f.Collector,
	}

	return f
}

// RunScenario loads a scenario file and runs it as a subtest.
func (f *Framework) RunScenario(t *testing.T, path string) {
	t.Helper()

	s, err := scenario.Load(path)
	if err != nil {
		t.Fatalf("loading scenario %s: %v", path, err)
	}

	t.Run(s.Name, func(t *testing.T) {
		result := f.Engine.Run(context.Background(), s)

		if !result.Passed {
			t.Errorf("scenario %q failed after %s: %v", s.Name, result.Duration, result.Error)
			if result.Diagnostics != nil {
				logDiagnostics(t, result.Diagnostics)
			}
		} else {
			t.Logf("scenario %q passed in %s", s.Name, result.Duration)
		}
	})
}

// RunScenarioDir loads all scenarios from a directory and runs each as a subtest.
func (f *Framework) RunScenarioDir(t *testing.T, dir string) {
	t.Helper()

	scenarios, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatalf("loading scenarios from %s: %v", dir, err)
	}

	for _, s := range scenarios {
		s := s
		t.Run(s.Name, func(t *testing.T) {
			result := f.Engine.Run(context.Background(), s)

			if !result.Passed {
				t.Errorf("scenario %q failed after %s: %v", s.Name, result.Duration, result.Error)
				if result.Diagnostics != nil {
					logDiagnostics(t, result.Diagnostics)
				}
			} else {
				t.Logf("scenario %q passed in %s", s.Name, result.Duration)
			}
		})
	}
}

func logDiagnostics(t *testing.T, report *diagnostics.Report) {
	t.Helper()

	if report.CollectionError != "" {
		t.Logf("diagnostics collection error: %s", report.CollectionError)
	}
	for name, logs := range report.AgentLogs {
		t.Logf("--- agent %s logs ---\n%s", name, logs)
	}
	for _, event := range report.Events {
		t.Logf("event: %s", event)
	}
	for _, diff := range report.ResourceDiffs {
		t.Logf("resource diff: %s %s: expected=%v actual=%v",
			diff.Resource, diff.Path, diff.Expected, diff.Actual)
	}
}
