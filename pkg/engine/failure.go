package engine

import (
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// FailureContext captures state at the point of scenario failure for diagnostics.
type FailureContext struct {
	Cluster          cluster.Cluster
	Scenario         *scenario.Scenario
	ScenarioPath     string
	SandboxNamespace string
	ArtifactsDir     string
	StartedAt        time.Time
	FailedAt         time.Time
	LastObserved     []ObservedCondition
	WatchLog         []WatchEvent
	AssertionIndex   int
}
