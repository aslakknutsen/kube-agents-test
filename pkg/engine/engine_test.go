package engine_test

import (
	"errors"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

func TestInfrastructureError(t *testing.T) {
	root := errors.New("root")
	err := &engine.InfrastructureError{Phase: "cluster", Err: root}
	if !errors.Is(err, root) {
		t.Fatal("expected unwrap to root")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error string")
	}
}
