// Engine example: call ScenarioEngine.Execute with explicit dependencies.
//
// Run: go run ./examples/engine
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/fault"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func main() {
	ctx := context.Background()

	sc := &scenario.Scenario{
		Name:    "leader-failover",
		Agents:  []string{"leader-agent", "follower-agent"},
		Setup:   scenario.Setup{Manifests: []string{"fixtures/leader-setup.yaml"}},
		Trigger: &scenario.Trigger{
			Fault: &fault.Spec{
				Kind:  fault.KindKillAgent,
				Agent: "leader-agent",
			},
		},
		Expect: scenario.Expect{
			Timeout: 2 * time.Minute,
			Assertions: []scenario.ResourceAssertion{
				{
					Resource: scenario.ResourceRef{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "workload",
						Namespace:  "test",
					},
					Conditions: []scenario.PathCondition{
						{Path: ".status.readyReplicas", Value: 3},
					},
				},
			},
		},
	}
	if err := sc.Validate(); err != nil {
		log.Fatalf("Validate: %v", err)
	}

	eng := &demoEngine{}
	result, err := eng.Execute(ctx, engine.ExecuteRequest{
		Scenario:  sc,
		BaseDir:   "/path/to/scenarios/leader-failover",
		Cluster:   &demoHandle{},
		Agents:    &demoAgentManager{},
		Collector: &noopCollector{},
	})
	if err != nil {
		log.Fatalf("Execute returned error: %v", err)
	}

	fmt.Printf("scenario: %s\n", result.ScenarioName)
	fmt.Printf("status: %s\n", result.Status)
	fmt.Printf("duration: %s\n", result.Duration)
	if result.Failure != nil {
		fmt.Printf("failure reason: %s\n", result.Failure.Reason)
		fmt.Printf("mismatches: %d\n", len(result.Failure.Mismatches))
		if result.Failure.Diagnostics != nil {
			fmt.Printf("diagnostics events: %d\n", len(result.Failure.Diagnostics.Events))
		}
	}
}

type demoEngine struct{}

func (e *demoEngine) Execute(ctx context.Context, req engine.ExecuteRequest) (*engine.Result, error) {
	// Real engine: apply setup manifests, fire trigger, poll expect assertions.
	_ = req.Cluster
	_ = req.Agents
	_ = req.Collector
	return &engine.Result{
		ScenarioName: req.Scenario.Name,
		Status:       engine.StatusPassed,
		Duration:     100 * time.Millisecond,
	}, nil
}

type demoHandle struct{}

func (h *demoHandle) Kubeconfig() ([]byte, error) { return nil, nil }
func (h *demoHandle) Release(ctx context.Context) error { return nil }

var _ cluster.Handle = (*demoHandle)(nil)

type demoAgentManager struct{}

func (m *demoAgentManager) DeploySet(ctx context.Context, agents []string) error { return nil }
func (m *demoAgentManager) Start(ctx context.Context, name string) error         { return nil }
func (m *demoAgentManager) Stop(ctx context.Context, name string) error          { return nil }
func (m *demoAgentManager) Kill(ctx context.Context, name string) error          { return nil }
func (m *demoAgentManager) ApplyDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (m *demoAgentManager) ClearDegradation(ctx context.Context, deg agent.Degradation) error {
	return nil
}
func (m *demoAgentManager) Teardown(ctx context.Context) error { return nil }

var _ agent.Manager = (*demoAgentManager)(nil)

type noopCollector struct{}

func (c *noopCollector) Collect(ctx context.Context, req diagnostics.CollectRequest) (*diagnostics.Bundle, error) {
	return &diagnostics.Bundle{}, nil
}

var _ diagnostics.Collector = (*noopCollector)(nil)
