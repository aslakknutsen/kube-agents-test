package cluster

// Mode selects ephemeral or attached cluster provisioning.
type Mode int

const (
	ModeEphemeral Mode = iota
	ModeAttached
)

// Config describes how a cluster should be provisioned or attached.
type Config struct {
	Mode      Mode
	Ephemeral *EphemeralConfig
	Attached  *AttachedConfig
}

// EphemeralConfig configures a provider-created local cluster.
type EphemeralConfig struct {
	Backend     string
	ClusterName string
}

// AttachedConfig connects to an existing cluster via kubeconfig.
type AttachedConfig struct {
	KubeconfigPath string
	Context        string
	LeaveRunning   bool
}
