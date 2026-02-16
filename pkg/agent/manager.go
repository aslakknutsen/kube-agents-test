package agent

import "context"

// DeployMode controls how an agent is run during a test.
type DeployMode string

const (
	// DeployModePod runs the agent as a Kubernetes pod (production-like).
	DeployModePod DeployMode = "pod"

	// DeployModeLocal runs the agent as a local process (faster iteration).
	DeployModeLocal DeployMode = "local"
)

// AgentConfig describes a single agent to manage during a test.
type AgentConfig struct {
	Name       string
	Image      string     // container image (used when Mode == DeployModeLocal is false)
	BinaryPath string     // local binary path (used when Mode == DeployModeLocal)
	Mode       DeployMode
	Namespace  string
}

// Manager controls the lifecycle of agents within a test cluster.
type Manager interface {
	// Deploy starts the given agents in the cluster.
	Deploy(ctx context.Context, agents []AgentConfig) error

	// Stop stops a specific agent by name.
	Stop(ctx context.Context, agentName string) error

	// Start restarts a previously stopped agent.
	Start(ctx context.Context, agentName string) error

	// StopAll stops all managed agents.
	StopAll(ctx context.Context) error

	// Logs returns the logs for a given agent.
	Logs(ctx context.Context, agentName string) (string, error)
}

// DefaultManager is a stub implementation of Manager.
type DefaultManager struct {
	Kubeconfig string
	agents     map[string]AgentConfig
}

func NewDefaultManager(kubeconfig string) *DefaultManager {
	return &DefaultManager{
		Kubeconfig: kubeconfig,
		agents:     make(map[string]AgentConfig),
	}
}

func (m *DefaultManager) Deploy(ctx context.Context, agents []AgentConfig) error {
	for _, a := range agents {
		m.agents[a.Name] = a
	}
	// TODO: actually deploy agents as pods or local processes
	return nil
}

func (m *DefaultManager) Stop(ctx context.Context, agentName string) error {
	// TODO: delete pod or kill process
	return nil
}

func (m *DefaultManager) Start(ctx context.Context, agentName string) error {
	// TODO: recreate pod or restart process
	return nil
}

func (m *DefaultManager) StopAll(ctx context.Context) error {
	// TODO: stop all agents
	return nil
}

func (m *DefaultManager) Logs(ctx context.Context, agentName string) (string, error) {
	// TODO: fetch pod logs or read process stdout
	return "", nil
}
