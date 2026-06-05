package cluster

import (
	"context"

	"k8s.io/client-go/rest"
)

// Mode selects ephemeral provisioning or an attached kubeconfig.
type Mode int

const (
	// Ephemeral creates a local cluster (e.g. kind) for the test run.
	Ephemeral Mode = iota
	// Attached uses an operator-supplied kubeconfig.
	Attached
)

// Config describes how to obtain a cluster for a test suite.
type Config struct {
	Mode           Mode
	KubeconfigPath string
	Ephemeral      *EphemeralConfig
}

// EphemeralConfig holds opaque settings for ephemeral cluster backends.
type EphemeralConfig struct {
	ClusterName string
	KindProfile string
}

// Handle is a live cluster connection returned by Provider.Ensure.
type Handle struct {
	KubeconfigPath string
	ClusterID      string
	RESTConfig     *rest.Config
}

// Provider provisions and tears down clusters.
type Provider interface {
	Ensure(ctx context.Context, cfg Config) (*Handle, error)
	Teardown(ctx context.Context, handle *Handle) error
}
