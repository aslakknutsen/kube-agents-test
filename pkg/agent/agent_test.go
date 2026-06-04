package agent_test

import (
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
)

func TestParseDeployMode(t *testing.T) {
	tests := []struct {
		in   string
		want agent.DeployMode
	}{
		{"pods", agent.DeployPods},
		{"", agent.DeployPods},
		{"local", agent.DeployLocalProcess},
	}
	for _, tc := range tests {
		got, err := agent.ParseDeployMode(tc.in)
		if err != nil {
			t.Fatalf("ParseDeployMode(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseDeployMode(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	if _, err := agent.ParseDeployMode("bogus"); err == nil {
		t.Fatal("expected error for bogus deploy mode")
	}
}
