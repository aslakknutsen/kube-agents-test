package runner

import "context"

// Runner orchestrates cluster, agent, engine, and diagnostics subsystems.
type Runner interface {
	Run(ctx context.Context, opts RunOptions) (Report, error)
}

// NewDefault returns a runner with the given subsystem dependencies.
func NewDefault(deps Dependencies) Runner {
	return &defaultRunner{deps: deps}
}
