// Package fault defines typed fault specifications for scenario triggers.
// See docs/fault-injection.md for the fault catalog.
package fault

import "time"

// Kind identifies a fault injection mechanism.
type Kind string

const (
	KindKillAgent        Kind = "killAgent"
	KindNetworkPartition Kind = "networkPartition"
	KindSlowAPIServer    Kind = "slowAPIServer"
	KindStaleCache       Kind = "staleCache"
	KindResourceConflict Kind = "resourceConflict"
)

// ResourcePatch applies a concurrent harness update for resourceConflict faults.
type ResourcePatch struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Name       string                 `yaml:"name"`
	Namespace  string                 `yaml:"namespace"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
}

// Spec describes a fault to inject during a scenario run.
// Additional fields will be added as the YAML schema is finalized.
type Spec struct {
	Kind  Kind   `yaml:"kind"`
	Agent string `yaml:"agent,omitempty"`

	Patch *ResourcePatch `yaml:"patch,omitempty"` // resourceConflict
	Delay time.Duration  `yaml:"delay,omitempty"` // slowAPIServer
}
