package diag_test

import (
	"context"
	"testing"

	"gitea.gitea/mirror/kube-agents-test/pkg/diag"
)

func TestNoopCollectorReturnsNilBundle(t *testing.T) {
	var c diag.NoopCollector
	b, err := c.Collect(context.Background(), diag.CollectRequest{})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if b != nil {
		t.Fatalf("expected nil bundle, got %#v", b)
	}
}

func TestNoopCollectorIdempotent(t *testing.T) {
	c := diag.NoopCollector{}
	for i := 0; i < 3; i++ {
		b, err := c.Collect(context.Background(), diag.CollectRequest{AgentIDs: []string{"x"}})
		if err != nil || b != nil {
			t.Fatalf("iteration %d: bundle=%#v err=%v", i, b, err)
		}
	}
}
