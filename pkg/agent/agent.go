// Package agent defines the Agent Manager API for deploying and controlling agents.
// See docs/architecture.md#agent-manager and docs/fault-injection.md.
package agent

import "context"

// DeploymentMode selects pod-based or local-process agent deployment.
type DeploymentMode string

const (
	ModePod          DeploymentMode = "pod"
	ModeLocalProcess DeploymentMode = "localProcess"
)

// Config holds manager-level settings. Per-agent image/command mapping is supplied
// via AgentRegistry (catalog format is defined in a follow-up issue).
type Config struct {
	Mode     DeploymentMode `yaml:"mode"`
	Registry AgentRegistry  `yaml:"-"`
}

// AgentRegistry resolves scenario agent names to deployable specs.
type AgentRegistry interface {
	Resolve(name string) (Spec, error)
}

// Spec describes how to deploy a single agent.
type Spec struct {
	Name    string
	Image   string   // pod mode
	Command []string // local process mode
}

// DegradationType identifies simulated degraded conditions.
type DegradationType string

const (
	DegradationResourceLimits DegradationType = "resourceLimits"
	DegradationNetworkPolicy  DegradationType = "networkPolicy"
)

// Degradation applies a degraded condition to an agent; type-specific fields live in internal/agent.
type Degradation struct {
	Agent string
	Type  DegradationType
}

// Manager controls Agent Set deployment and fault-related lifecycle.
type Manager interface {
	DeploySet(ctx context.Context, agents []string) error
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Kill(ctx context.Context, name string) error
	ApplyDegradation(ctx context.Context, deg Degradation) error
	ClearDegradation(ctx context.Context, deg Degradation) error
	Teardown(ctx context.Context) error
}
