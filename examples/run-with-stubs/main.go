// Run a scenario with stub backends — fails at cluster Ensure with ErrNotImplemented.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/kube-agents/kube-agents-test/internal/errs"
	"github.com/kube-agents/kube-agents-test/internal/stub"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/runner"
)

func main() {
	r := runner.NewDefault(runner.Dependencies{
		Provider:  stub.NewProvider(),
		Manager:   stub.NewManager(),
		Engine:    stub.NewEngine(),
		Collector: stub.NewCollector(),
	})

	report, err := r.Run(context.Background(), runner.RunOptions{
		Paths: []string{"examples/scenarios/scaling-quota/scenario.yaml"},
		ClusterConfig: cluster.Config{
			Mode: cluster.ModeEphemeral,
			Ephemeral: &cluster.EphemeralConfig{
				Backend:     "kind",
				ClusterName: "example",
			},
		},
		SandboxNamespace: "test-agents",
	})

	if errors.Is(err, errs.ErrNotImplemented) {
		fmt.Println("expected: cluster Ensure returned ErrNotImplemented (stub provider)")
		fmt.Printf("exit code: %d (infra failure)\n", report.ExitCode())
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("passed: %v\n", report.Passed())
	os.Exit(report.ExitCode())
}
