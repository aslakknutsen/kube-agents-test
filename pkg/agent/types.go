package agent

import (
	"context"

	"gitea.gitea/mirror/kube-agents-test/pkg/cluster"
)

// DeploymentMode selects pod-based or local-process agent execution.
type DeploymentMode int

const (
	// Pods runs agents as workloads in the test cluster.
	Pods DeploymentMode = iota
	// LocalProcess runs agents on the host with cluster credentials.
	LocalProcess
)

// AgentSpec describes how to run one agent.
type AgentSpec struct {
	ID     string
	Image  string
	Binary string
}

// DegradationSpec describes resource or network constraints on an agent.
type DegradationSpec struct {
	AgentID string
}

// Registry resolves scenario agent IDs to deployment specs.
type Registry interface {
	Resolve(ids []string) ([]AgentSpec, error)
}

// Manager controls agent lifecycle during a test run.
type Manager interface {
	DeploySet(ctx context.Context, handle *cluster.Handle, agents []string, mode DeploymentMode) error
	Start(ctx context.Context, agentID string) error
	Stop(ctx context.Context, agentID string) error
	Kill(ctx context.Context, agentID string) error
	ApplyDegradation(ctx context.Context, spec DegradationSpec) error
	Teardown(ctx context.Context) error
}
