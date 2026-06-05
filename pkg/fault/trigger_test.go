package fault_test

import (
	"errors"
	"testing"
	"time"

	"gitea.gitea/mirror/kube-agents-test/pkg/fault"
)

func TestTriggerValidate(t *testing.T) {
	t.Run("patch ok", func(t *testing.T) {
		tr := fault.Trigger{Patch: &fault.PatchTrigger{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "x",
		}}
		if err := tr.Validate(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("multiple variants", func(t *testing.T) {
		tr := fault.Trigger{
			Patch:        &fault.PatchTrigger{APIVersion: "v1", Kind: "Pod", Name: "p"},
			AgentRestart: &fault.AgentRestartTrigger{AgentID: "a"},
		}
		err := tr.Validate()
		if !errors.Is(err, fault.ErrInvalidTrigger) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("fault killAgent", func(t *testing.T) {
		tr := fault.Trigger{Fault: &fault.Fault{KillAgent: &fault.KillAgentFault{AgentID: "leader"}}}
		if err := tr.Validate(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestFaultValidate(t *testing.T) {
	f := fault.Fault{SlowAPIServer: &fault.SlowAPIServerFault{Latency: 100 * time.Millisecond}}
	if err := f.Validate(); err != nil {
		t.Fatal(err)
	}

	f2 := fault.Fault{SlowAPIServer: &fault.SlowAPIServerFault{}}
	if err := f2.Validate(); !errors.Is(err, fault.ErrInvalidTrigger) {
		t.Fatalf("got %v", err)
	}
}
