package engine

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// RunInput carries everything needed to execute one scenario.
type RunInput struct {
	Cluster          cluster.Cluster
	Scenario         *scenario.Scenario
	Manager          agent.Manager
	ScenarioPath     string
	SandboxNamespace string
}

// Engine executes a single scenario lifecycle: setup, trigger, and assertions.
type Engine interface {
	Run(ctx context.Context, in RunInput) (Result, error)
}
