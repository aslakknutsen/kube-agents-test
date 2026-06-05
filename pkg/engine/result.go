package engine

import (
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
	"gopkg.in/yaml.v3"
)

// Result is the outcome of executing a single scenario.
type Result struct {
	ScenarioName  string
	Passed        bool
	Duration      time.Duration
	FailureReason string
	Failure       *FailureContext
}

// ObservedCondition records an expected vs actual value at failure time.
type ObservedCondition struct {
	Path     string
	Expected yaml.Node
	Actual   yaml.Node
	Resource scenario.ResourceSelector
}

// WatchEvent records a resource mutation observed during polling.
type WatchEvent struct {
	Timestamp time.Time
	Resource  scenario.ResourceSelector
	Verb      string
}
