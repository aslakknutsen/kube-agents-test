package agent_test

import (
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
)

func TestParseDeployMode(t *testing.T) {
	got, err := agent.ParseDeployMode("local")
	if err != nil {
		t.Fatal(err)
	}
	if got != agent.DeployLocalProcess {
		t.Errorf("got %v", got)
	}
}
