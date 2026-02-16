// Package agent defines the interface for managing agents during tests.
package agent

import "context"

// Manager controls the lifecycle of agents within a test cluster.
type Manager interface {
	// Deploy starts the named agent in the cluster.
	Deploy(ctx context.Context, agentName string) error

	// Stop stops the named agent. It should be restartable via Deploy.
	Stop(ctx context.Context, agentName string) error

	// StopAll stops every deployed agent.
	StopAll(ctx context.Context) error

	// Logs returns the logs for the named agent since the last Deploy call.
	Logs(ctx context.Context, agentName string) (string, error)
}
