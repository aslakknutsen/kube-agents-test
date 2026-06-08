package runner

import (
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

// Dependencies injects subsystem implementations into the runner.
type Dependencies struct {
	Provider  cluster.Provider
	Manager   agent.Manager
	Engine    engine.Engine
	Collector diagnostics.Collector
}

// RunOptions configures a test run.
type RunOptions struct {
	Paths              []string
	ClusterConfig      cluster.Config
	AgentConfig        agent.Config
	Deps               Dependencies
	FailFast           bool
	LeaveCluster       bool
	SandboxNamespace   string
	AllowClusterScoped bool
	AllowFaults        bool
	ArtifactsDir       string
}
