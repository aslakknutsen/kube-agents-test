package fault_test

import (
	"errors"
	"testing"

	"gitea.gitea/mirror/kube-agents-test/pkg/fault"
	"gopkg.in/yaml.v3"
)

func TestFaultValidateRejectsAmbiguousYAML(t *testing.T) {
	t.Run("dual_variants_unmarshaled", func(t *testing.T) {
		var f fault.Fault
		const y = `killAgent: {agentId: a}
networkPartition: {agentId: b}
`
		if err := yaml.Unmarshal([]byte(y), &f); err != nil {
			t.Fatal(err)
		}
		err := f.Validate()
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !errors.Is(err, fault.ErrInvalidTrigger) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("empty_fault_object", func(t *testing.T) {
		tr := fault.Trigger{Fault: &fault.Fault{}}
		if err := tr.Validate(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("nil_trigger_pointer", func(t *testing.T) {
		var tr *fault.Trigger
		if err := tr.Validate(); !errors.Is(err, fault.ErrInvalidTrigger) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestSlowAPIServerRejectsZeroLatency(t *testing.T) {
	f := fault.Fault{SlowAPIServer: &fault.SlowAPIServerFault{Latency: 0}}
	err := f.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, fault.ErrInvalidTrigger) {
		t.Fatalf("got %v", err)
	}
}

func TestTriggerValidateRejectsWhitespaceAgentIDs(t *testing.T) {
	cases := []struct {
		name string
		tr   fault.Trigger
	}{
		{
			name: "agentRestart_whitespace",
			tr:   fault.Trigger{AgentRestart: &fault.AgentRestartTrigger{AgentID: "   "}},
		},
		{
			name: "agentRestart_nbsp",
			tr:   fault.Trigger{AgentRestart: &fault.AgentRestartTrigger{AgentID: "\u00a0"}},
		},
		{
			name: "killAgent_whitespace",
			tr:   fault.Trigger{Fault: &fault.Fault{KillAgent: &fault.KillAgentFault{AgentID: "   "}}},
		},
		{
			name: "networkPartition_whitespace",
			tr:   fault.Trigger{Fault: &fault.Fault{NetworkPartition: &fault.NetworkPartitionFault{AgentID: "   "}}},
		},
		{
			name: "staleCache_whitespace",
			tr:   fault.Trigger{Fault: &fault.Fault{StaleCache: &fault.StaleCacheFault{AgentID: "   "}}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.tr.Validate(); !errors.Is(err, fault.ErrInvalidTrigger) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestTriggerValidateRejectsWhitespacePatchFields(t *testing.T) {
	tr := fault.Trigger{Patch: &fault.PatchTrigger{
		APIVersion: "   ",
		Kind:       "Pod",
		Name:       "p",
	}}
	if err := tr.Validate(); !errors.Is(err, fault.ErrInvalidTrigger) {
		t.Fatalf("got %v", err)
	}
}

func TestFaultValidateRejectsIncompleteVariants(t *testing.T) {
	cases := []struct {
		name string
		f    fault.Fault
	}{
		{
			name: "killAgent_missing_id",
			f:    fault.Fault{KillAgent: &fault.KillAgentFault{}},
		},
		{
			name: "networkPartition_missing_id",
			f:    fault.Fault{NetworkPartition: &fault.NetworkPartitionFault{}},
		},
		{
			name: "staleCache_missing_id",
			f:    fault.Fault{StaleCache: &fault.StaleCacheFault{}},
		},
		{
			name: "resourceConflict_missing_kind",
			f:    fault.Fault{ResourceConflict: &fault.ResourceConflictFault{APIVersion: "v1", Name: "x"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.f.Validate(); !errors.Is(err, fault.ErrInvalidTrigger) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
