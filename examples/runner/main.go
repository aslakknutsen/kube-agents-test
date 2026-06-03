// Runner example: wire cluster, agent, and engine dependencies and run scenarios.
//
// Run: go run ./examples/runner
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

const sampleScenario = `name: minimal-example
agents:
  - scaling-agent
setup:
  manifests:
    - fixtures/setup.yaml
expect:
  assertions:
    - resource:
        apiVersion: v1
        kind: ConfigMap
        name: cfg
        namespace: test
      conditions:
        - path: .metadata.name
          value: cfg
  timeout: 30s
`

func main() {
	ctx := context.Background()

	dir, err := os.MkdirTemp("", "kube-agents-run-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	scenarioPath := filepath.Join(dir, "minimal.yaml")
	if err := os.WriteFile(scenarioPath, []byte(sampleScenario), 0o644); err != nil {
		log.Fatal(err)
	}

	// Production code would use kind/attached providers and a real engine; fakes
	// show the dependency injection shape without a cluster.
	r := runner.New(runner.Config{
		Cluster: cluster.Config{
			Mode: cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{
				Provider: "kind",
			},
		},
		Agent: agent.Config{
			Mode: agent.ModePod,
			Registry: &memoryRegistry{
				specs: map[string]agent.Spec{
					"scaling-agent": {Name: "scaling-agent", Image: "example/scaling:latest"},
				},
			},
		},
	}, runner.Dependencies{
		ClusterProvider: &fakeClusterProvider{},
		AgentManager:    &fakeAgentManager{},
		Engine:          &fakeEngine{},
	})

	// Run a single loaded scenario.
	sc, err := scenario.Load(scenarioPath)
	if err != nil {
		log.Fatalf("Load: %v", err)
	}
	report, err := r.Run(ctx, []*scenario.Scenario{sc})
	if err != nil {
		log.Fatalf("Run: %v", err)
	}
	printReport("Run", report)

	// Or pass a file or directory path (loads all .yaml/.yml scenarios).
	report, err = r.RunPath(ctx, dir)
	if err != nil {
		log.Fatalf("RunPath: %v", err)
	}
	printReport("RunPath", report)
}

func printReport(label string, report *runner.RunReport) {
	fmt.Printf("\n%s summary: total=%d passed=%d failed=%d errored=%d\n",
		label, report.Summary.Total, report.Summary.Passed,
		report.Summary.Failed, report.Summary.Errored)
	for _, res := range report.Results {
		fmt.Printf("  - %s: %s", res.ScenarioName, res.Status)
		if res.Duration > 0 {
			fmt.Printf(" (%s)", res.Duration)
		}
		if res.Failure != nil {
			fmt.Printf(" reason=%s mismatches=%d", res.Failure.Reason, len(res.Failure.Mismatches))
		}
		if res.Err != nil {
			fmt.Printf(" err=%v", res.Err)
		}
		fmt.Println()
	}
}

// --- fakes (replace with internal/* implementations) ---

type memoryRegistry struct {
	specs map[string]agent.Spec
}

func (r *memoryRegistry) Resolve(name string) (agent.Spec, error) {
	spec, ok := r.specs[name]
	if !ok {
		return agent.Spec{}, fmt.Errorf("unknown agent: %s", name)
	}
	return spec, nil
}

type fakeClusterProvider struct{}

func (p *fakeClusterProvider) Acquire(ctx context.Context) (cluster.Handle, error) {
	return &fakeHandle{}, nil
}

type fakeHandle struct{}

func (h *fakeHandle) Kubeconfig() ([]byte, error) {
	return []byte("apiVersion: v1\nkind: Config\n"), nil
}

func (h *fakeHandle) Release(ctx context.Context) error { return nil }

type fakeAgentManager struct{}

func (m *fakeAgentManager) DeploySet(ctx context.Context, agents []string) error { return nil }
func (m *fakeAgentManager) Start(ctx context.Context, name string) error         { return nil }
func (m *fakeAgentManager) Stop(ctx context.Context, name string) error          { return nil }
func (m *fakeAgentManager) Kill(ctx context.Context, name string) error          { return nil }
func (m *fakeAgentManager) ApplyDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (m *fakeAgentManager) ClearDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (m *fakeAgentManager) Teardown(ctx context.Context) error { return nil }

type fakeEngine struct{}

func (e *fakeEngine) Execute(ctx context.Context, req engine.ExecuteRequest) (*engine.Result, error) {
	return &engine.Result{
		ScenarioName: req.Scenario.Name,
		Status:       engine.StatusPassed,
		Duration:     50 * time.Millisecond,
	}, nil
}
