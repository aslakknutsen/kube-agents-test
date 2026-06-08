package agent

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Manager controls agent lifecycle and degradation hooks. It never parses YAML.
type Manager interface {
	Deploy(ctx context.Context, c cluster.Cluster, agents scenario.AgentSet) error
	Start(ctx context.Context, agentID string) error
	Stop(ctx context.Context, agentID string) error
	Kill(ctx context.Context, agentID string) error
	Teardown(ctx context.Context) error

	SetResourceLimits(ctx context.Context, agentID string, limits ResourceLimits) error
	ClearResourceLimits(ctx context.Context, agentID string) error
	ApplyNetworkPolicy(ctx context.Context, spec NetworkPolicySpec) (string, error)
	RemoveNetworkPolicy(ctx context.Context, policyID string) error
}
