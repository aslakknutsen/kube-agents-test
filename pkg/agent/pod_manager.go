package agent

import (
	"context"
	"fmt"
)

// AgentConfig holds the configuration for deploying a single agent.
type AgentConfig struct {
	// Name is the agent identifier (must match the name used in scenario YAML).
	Name string

	// Image is the container image to deploy (e.g. "ghcr.io/kube-agents/scaling-agent:latest").
	Image string

	// Namespace is the Kubernetes namespace to deploy into.
	Namespace string

	// Args are additional command-line arguments passed to the agent container.
	Args []string
}

// PodManager deploys agents as Kubernetes pods in the test cluster.
type PodManager struct {
	// KubeconfigPath is the path to the cluster kubeconfig.
	KubeconfigPath string

	// Agents maps agent names to their deployment configuration.
	Agents map[string]AgentConfig

	deployed map[string]bool
}

// NewPodManager creates a PodManager with the given kubeconfig and agent configs.
func NewPodManager(kubeconfigPath string, agents []AgentConfig) *PodManager {
	m := &PodManager{
		KubeconfigPath: kubeconfigPath,
		Agents:         make(map[string]AgentConfig),
		deployed:       make(map[string]bool),
	}
	for _, a := range agents {
		m.Agents[a.Name] = a
	}
	return m
}

func (m *PodManager) Deploy(ctx context.Context, agentName string) error {
	cfg, ok := m.Agents[agentName]
	if !ok {
		return fmt.Errorf("unknown agent %q", agentName)
	}

	// TODO: Use client-go to create a Pod/Deployment for the agent.
	_ = cfg
	m.deployed[agentName] = true
	return nil
}

func (m *PodManager) Stop(ctx context.Context, agentName string) error {
	if !m.deployed[agentName] {
		return fmt.Errorf("agent %q is not deployed", agentName)
	}

	// TODO: Use client-go to delete the agent's Pod/Deployment.
	delete(m.deployed, agentName)
	return nil
}

func (m *PodManager) StopAll(ctx context.Context) error {
	for name := range m.deployed {
		if err := m.Stop(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (m *PodManager) Logs(ctx context.Context, agentName string) (string, error) {
	// TODO: Use client-go to fetch pod logs.
	return "", fmt.Errorf("not implemented")
}
