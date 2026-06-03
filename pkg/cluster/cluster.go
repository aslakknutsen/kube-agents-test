// Package cluster defines the Cluster Provider API for acquiring test cluster access.
// See docs/architecture.md#cluster-provider.
package cluster

import "context"

// Mode selects ephemeral (kind/k3d) or attached (existing kubeconfig) clusters.
type Mode string

const (
	ModeEphemeral Mode = "ephemeral"
	ModeAttached  Mode = "attached"
)

// Config holds cluster provider settings for a test run.
type Config struct {
	Mode      Mode              `yaml:"mode"`
	Ephemeral *EphemeralConfig  `yaml:"ephemeral,omitempty"`
	Attached  *AttachedConfig   `yaml:"attached,omitempty"`
}

// EphemeralConfig configures a throwaway local cluster (kind first, k3d later).
type EphemeralConfig struct {
	Provider string `yaml:"provider"` // "kind" first; "k3d" later
}

// AttachedConfig connects to an existing cluster via kubeconfig.
type AttachedConfig struct {
	KubeconfigPath string `yaml:"kubeconfigPath"`
}

// Provider acquires cluster access for a test run.
type Provider interface {
	Acquire(ctx context.Context) (Handle, error)
}

// Handle is active cluster access consumed by the Agent Manager and Scenario Engine.
type Handle interface {
	// Kubeconfig returns kubeconfig bytes for the active cluster.
	Kubeconfig() ([]byte, error)
	// Release tears down ephemeral clusters or releases attached handles.
	Release(ctx context.Context) error
}
