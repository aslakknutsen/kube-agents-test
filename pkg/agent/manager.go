package agent

import "context"

// Manager handles deploying, starting, stopping, and killing agents
// within the test cluster.
type Manager interface {
	// Deploy installs the named agent into the cluster.
	Deploy(ctx context.Context, agentName string) error

	// Start starts a previously deployed agent.
	Start(ctx context.Context, agentName string) error

	// Stop gracefully stops an agent.
	Stop(ctx context.Context, agentName string) error

	// Kill forcefully terminates an agent (simulates crash).
	Kill(ctx context.Context, agentName string) error

	// Logs returns recent log output for the named agent.
	Logs(ctx context.Context, agentName string) (string, error)
}

// DeployMode controls how agents are run during tests.
type DeployMode int

const (
	// DeployModePod runs agents as pods in the cluster (production-like).
	DeployModePod DeployMode = iota
	// DeployModeLocal runs agents as local processes (faster iteration).
	DeployModeLocal
)

// AgentConfig describes how to deploy a single agent.
type AgentConfig struct {
	Name       string
	Image      string
	DeployMode DeployMode
	Command    []string
	Args       []string
}

// PodManager deploys agents as Kubernetes pods.
type PodManager struct {
	kubeconfig string
	configs    map[string]AgentConfig
}

func NewPodManager(kubeconfig string) *PodManager {
	return &PodManager{
		kubeconfig: kubeconfig,
		configs:    make(map[string]AgentConfig),
	}
}

func (m *PodManager) RegisterAgent(cfg AgentConfig) {
	m.configs[cfg.Name] = cfg
}

func (m *PodManager) Deploy(ctx context.Context, agentName string) error {
	// TODO: create pod/deployment for the agent using client-go
	return nil
}

func (m *PodManager) Start(ctx context.Context, agentName string) error {
	// TODO: scale deployment to 1
	return nil
}

func (m *PodManager) Stop(ctx context.Context, agentName string) error {
	// TODO: scale deployment to 0
	return nil
}

func (m *PodManager) Kill(ctx context.Context, agentName string) error {
	// TODO: delete pod with grace period 0
	return nil
}

func (m *PodManager) Logs(ctx context.Context, agentName string) (string, error) {
	// TODO: fetch pod logs via client-go
	return "", nil
}
