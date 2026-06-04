package scenario_test

import (
	"path/filepath"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadDir(t *testing.T) {
	dir := filepath.Join("testdata")
	scenarios, err := scenario.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(scenarios) < 2 {
		t.Fatalf("expected at least 2 scenarios, got %d", len(scenarios))
	}
}
