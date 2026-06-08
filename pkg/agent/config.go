package agent

// DeploymentMode selects how agents are run in a test cluster.
type DeploymentMode int

const (
	DeployPods DeploymentMode = iota
	DeployLocalProcess
)

// Config configures agent deployment for a test run.
type Config struct {
	Mode      DeploymentMode
	Namespace string
	Images    map[string]string
	Binaries  map[string]string
}

// ResourceLimits describes CPU and memory constraints for an agent.
type ResourceLimits struct {
	CPU    string
	Memory string
}

// NetworkPolicySpec describes a network policy applied for fault injection.
type NetworkPolicySpec struct {
	Namespace string
	AgentID   string
}
