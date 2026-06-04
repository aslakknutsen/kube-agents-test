// Package cluster defines the Cluster Provider API for ephemeral and attached modes.
//
// API v0 — unstable until first release.
package cluster

import (
	"context"

	"k8s.io/client-go/rest"
)

// Mode selects ephemeral (provisioned) or attached (existing kubeconfig) clusters.
type Mode int

const (
	ModeEphemeral Mode = iota
	ModeAttached
)

// String returns a stable name for Mode.
func (m Mode) String() string {
	switch m {
	case ModeEphemeral:
		return "ephemeral"
	case ModeAttached:
		return "attached"
	default:
		return "unknown"
	}
}

// ParseMode parses ephemeral|attached (case-insensitive).
func ParseMode(s string) (Mode, error) {
	switch s {
	case "ephemeral", "":
		return ModeEphemeral, nil
	case "attached":
		return ModeAttached, nil
	default:
		return 0, ErrInvalidMode
	}
}

// Cluster holds kubeconfig path and REST config after provision or attach.
type Cluster struct {
	KubeconfigPath string
	RESTConfig     *rest.Config
}

// EphemeralOptions configures ephemeral cluster backends (e.g. kind).
type EphemeralOptions struct {
	Backend     string // "kind" is the default per roadmap
	ClusterName string
}

// Provider supplies clusters for test runs.
type Provider interface {
	Provision(ctx context.Context) (*Cluster, error)
	Attach(ctx context.Context, kubeconfig string) (*Cluster, error)
	// Teardown releases ephemeral clusters; attached mode is a no-op or best-effort.
	Teardown(ctx context.Context) error
}
