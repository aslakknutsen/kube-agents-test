package runner

import (
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
)

// Report aggregates per-scenario results for a run.
type Report struct {
	Results           []engine.Result
	Diagnostics       map[string]diagnostics.Artifacts
	DiagnosticsErrors map[string]error
	StartedAt         time.Time
	EndedAt           time.Time
	InfraFailed       bool
}

// Passed reports whether every scenario passed.
func (r Report) Passed() bool {
	for _, result := range r.Results {
		if !result.Passed {
			return false
		}
	}
	return len(r.Results) > 0
}

// FailedCount returns the number of failed scenarios.
func (r Report) FailedCount() int {
	count := 0
	for _, result := range r.Results {
		if !result.Passed {
			count++
		}
	}
	return count
}

// ExitCode returns 0 if all passed, 1 if any scenario failed, 2 for infrastructure errors.
func (r Report) ExitCode() int {
	if r.InfraFailed {
		return 2
	}
	if r.Passed() {
		return 0
	}
	return 1
}
