package main

import (
	"strings"
	"testing"
)

func TestRun_requiresRunSubcommand(t *testing.T) {
	err := run([]string{"deploy", "x"})
	if err == nil || !strings.Contains(err.Error(), "only 'run'") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunCommand_requiresSinglePath(t *testing.T) {
	err := runCommand([]string{"-cluster-mode=ephemeral"})
	if err == nil || !strings.Contains(err.Error(), "exactly one path") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunCommand_rejectsUnknownClusterMode(t *testing.T) {
	err := runCommand([]string{"-cluster-mode=bogus", "pkg/scenario/testdata"})
	if err == nil {
		t.Fatal("expected error for bogus cluster mode")
	}
}

func TestRunCommand_rejectsUnknownAgentMode(t *testing.T) {
	err := runCommand([]string{"-agent-mode=bogus", "pkg/scenario/testdata"})
	if err == nil {
		t.Fatal("expected error for bogus agent mode")
	}
}
