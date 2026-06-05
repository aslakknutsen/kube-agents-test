package fault

import (
	"fmt"
	"strings"
	"time"
)

// Trigger is a discriminated union of scenario perturbation types.
// Exactly one non-nil field is required when Trigger is present.
type Trigger struct {
	Patch        *PatchTrigger        `yaml:"patch,omitempty"`
	AgentRestart *AgentRestartTrigger `yaml:"agentRestart,omitempty"`
	Fault        *Fault               `yaml:"fault,omitempty"`
}

// PatchTrigger applies a resource mutation via patch body fields.
type PatchTrigger struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Body       map[string]interface{} `yaml:",inline"`
}

// AgentRestartTrigger restarts an agent mid-scenario.
type AgentRestartTrigger struct {
	AgentID     string        `yaml:"agentId"`
	GracePeriod time.Duration `yaml:"gracePeriod,omitempty"`
}

// Fault is a nested discriminant for fault-injection hooks.
type Fault struct {
	KillAgent         *KillAgentFault         `yaml:"killAgent,omitempty"`
	NetworkPartition  *NetworkPartitionFault  `yaml:"networkPartition,omitempty"`
	SlowAPIServer     *SlowAPIServerFault     `yaml:"slowAPIServer,omitempty"`
	StaleCache        *StaleCacheFault        `yaml:"staleCache,omitempty"`
	ResourceConflict  *ResourceConflictFault  `yaml:"resourceConflict,omitempty"`
}

// KillAgentFault kills an agent pod or process.
type KillAgentFault struct {
	AgentID string `yaml:"agentId"`
}

// NetworkPartitionFault isolates an agent from the API server.
type NetworkPartitionFault struct {
	AgentID   string `yaml:"agentId"`
	Namespace string `yaml:"namespace,omitempty"`
}

// SlowAPIServerFault injects API latency via proxy.
type SlowAPIServerFault struct {
	Latency time.Duration `yaml:"latency"`
}

// StaleCacheFault restarts an informer without full resync.
type StaleCacheFault struct {
	AgentID string `yaml:"agentId"`
}

// ResourceConflictFault applies a concurrent harness update.
type ResourceConflictFault struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Patch      map[string]interface{} `yaml:"patch,omitempty"`
}

// Validate checks that exactly one trigger variant is set and each variant is valid.
func (t *Trigger) Validate() error {
	if t == nil {
		return fmt.Errorf("%w: trigger is nil", ErrInvalidTrigger)
	}
	n := 0
	if t.Patch != nil {
		n++
	}
	if t.AgentRestart != nil {
		n++
	}
	if t.Fault != nil {
		n++
	}
	if n != 1 {
		return fmt.Errorf("%w: exactly one of patch, agentRestart, or fault must be set", ErrInvalidTrigger)
	}
	if t.Patch != nil {
		return t.Patch.validate()
	}
	if t.AgentRestart != nil {
		return t.AgentRestart.validate()
	}
	return t.Fault.Validate()
}

func (p *PatchTrigger) validate() error {
	if isBlank(p.APIVersion) || isBlank(p.Kind) || isBlank(p.Name) {
		return fmt.Errorf("%w: patch requires apiVersion, kind, and name", ErrInvalidTrigger)
	}
	return nil
}

func (a *AgentRestartTrigger) validate() error {
	if isBlank(a.AgentID) {
		return fmt.Errorf("%w: agentRestart requires agentId", ErrInvalidTrigger)
	}
	return nil
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Validate checks that exactly one fault variant is set and required fields are present.
func (f *Fault) Validate() error {
	if f == nil {
		return fmt.Errorf("%w: fault is nil", ErrInvalidTrigger)
	}
	n := 0
	if f.KillAgent != nil {
		n++
	}
	if f.NetworkPartition != nil {
		n++
	}
	if f.SlowAPIServer != nil {
		n++
	}
	if f.StaleCache != nil {
		n++
	}
	if f.ResourceConflict != nil {
		n++
	}
	if n != 1 {
		return fmt.Errorf("%w: exactly one fault variant must be set", ErrInvalidTrigger)
	}
	switch {
	case f.KillAgent != nil:
		if isBlank(f.KillAgent.AgentID) {
			return fmt.Errorf("%w: killAgent requires agentId", ErrInvalidTrigger)
		}
	case f.NetworkPartition != nil:
		if isBlank(f.NetworkPartition.AgentID) {
			return fmt.Errorf("%w: networkPartition requires agentId", ErrInvalidTrigger)
		}
	case f.SlowAPIServer != nil:
		if f.SlowAPIServer.Latency <= 0 {
			return fmt.Errorf("%w: slowAPIServer requires positive latency", ErrInvalidTrigger)
		}
	case f.StaleCache != nil:
		if isBlank(f.StaleCache.AgentID) {
			return fmt.Errorf("%w: staleCache requires agentId", ErrInvalidTrigger)
		}
	case f.ResourceConflict != nil:
		if isBlank(f.ResourceConflict.APIVersion) || isBlank(f.ResourceConflict.Kind) || isBlank(f.ResourceConflict.Name) {
			return fmt.Errorf("%w: resourceConflict requires apiVersion, kind, and name", ErrInvalidTrigger)
		}
	}
	return nil
}
