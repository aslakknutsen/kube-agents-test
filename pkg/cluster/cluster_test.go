package cluster_test

import (
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		in   string
		want cluster.Mode
	}{
		{"ephemeral", cluster.ModeEphemeral},
		{"attached", cluster.ModeAttached},
		{"", cluster.ModeEphemeral},
	}
	for _, tc := range tests {
		got, err := cluster.ParseMode(tc.in)
		if err != nil {
			t.Fatalf("ParseMode(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("ParseMode(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
	if _, err := cluster.ParseMode("bogus"); err == nil {
		t.Fatal("expected error for bogus mode")
	}
}
