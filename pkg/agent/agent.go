// Package agent defines the Agent Manager API for deploying and controlling agents.
//
// API v0 — unstable until first release.
package agent

import (
	"context"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

// DeployMode selects in-cluster pods or local host processes.
type DeployMode int

const (
	DeployPods DeployMode = iota
	DeployLocalProcess
)

// String returns a stable name for DeployMode.
func (m DeployMode) String() string {
	switch m {
	case DeployPods:
		return "pods"
	case DeployLocalProcess:
		return "local"
	default:
		return "unknown"
	}
}

// ParseDeployMode parses pods|local (case-insensitive).
func ParseDeployMode(s string) (DeployMode, error) {
	switch s {
	case "pods", "":
		return DeployPods, nil
	case "local":
		return DeployLocalProcess, nil
	default:
		return 0, ErrInvalidDeployMode
	}
}

// Config holds agent deployment settings external to scenario YAML.
type Config struct {
	Mode     DeployMode
	Registry map[string]AgentSpec
}

// AgentSpec maps an agent name to image and optional manifests.
type AgentSpec struct {
	Image     string
	Manifests []string
}

// DegradedOptions simulates constrained agents (limits, network policy).
type DegradedOptions struct {
	ResourceLimits  *ResourceLimitSpec
	NetworkPolicy   *NetworkPolicySpec
}

// ResourceLimitSpec is a placeholder for resource limit fault injection.
type ResourceLimitSpec struct {
	CPU    string
	Memory string
}

// NetworkPolicySpec is a placeholder for network policy fault injection.
type NetworkPolicySpec struct {
	DenyEgress bool
}

// Manager controls agent lifecycle within a test cluster.
type Manager interface {
	DeploySet(ctx context.Context, agents []string) error
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Kill(ctx context.Context, name string) error
	ApplyDegraded(ctx context.Context, name string, opts DegradedOptions) error
	Teardown(ctx context.Context) error
}

// Factory constructs a Manager bound to a cluster.
type Factory func(cluster *cluster.Cluster, cfg Config) (Manager, error)
