package runner

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/engine"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

type defaultRunner struct {
	deps Dependencies
}

func (r *defaultRunner) Run(ctx context.Context, opts RunOptions) (Report, error) {
	if opts.Deps.Provider == nil {
		opts.Deps = r.deps
	}
	if opts.Deps.Provider == nil || opts.Deps.Manager == nil || opts.Deps.Engine == nil {
		return Report{}, fmt.Errorf("runner dependencies are incomplete")
	}

	loaded, err := scenario.LoadPaths(opts.Paths)
	if err != nil {
		return Report{}, err
	}

	valCtx := scenario.ValidationContext{
		SandboxNamespace:   opts.SandboxNamespace,
		AllowClusterScoped: opts.AllowClusterScoped,
		AllowFaults:        opts.AllowFaults,
	}
	for _, item := range loaded {
		if err := item.Scenario.ValidateWith(valCtx); err != nil {
			return Report{}, err
		}
	}

	report := Report{
		StartedAt:         time.Now(),
		Diagnostics:       make(map[string]diagnostics.Artifacts),
		DiagnosticsErrors: make(map[string]error),
	}

	cl, err := opts.Deps.Provider.Ensure(ctx, opts.ClusterConfig)
	if err != nil {
		report.InfraFailed = true
		report.EndedAt = time.Now()
		return report, err
	}

	leaveCluster := opts.LeaveCluster
	if opts.ClusterConfig.Mode == cluster.ModeAttached && opts.ClusterConfig.Attached != nil {
		leaveCluster = leaveCluster || opts.ClusterConfig.Attached.LeaveRunning
	}
	if !leaveCluster {
		defer func() {
			_ = opts.Deps.Provider.Teardown(ctx, cl)
		}()
	}

	var deployedAgents scenario.AgentSet
	defer func() {
		if len(deployedAgents) > 0 {
			_ = opts.Deps.Manager.Teardown(ctx)
		}
	}()

	for _, item := range loaded {
		if !agentSetsEqual(item.Scenario.Agents, deployedAgents) {
			if len(deployedAgents) > 0 {
				if err := opts.Deps.Manager.Teardown(ctx); err != nil {
					report.InfraFailed = true
					report.EndedAt = time.Now()
					return report, err
				}
				deployedAgents = nil
			}
			if err := opts.Deps.Manager.Deploy(ctx, cl, item.Scenario.Agents); err != nil {
				report.InfraFailed = true
				report.EndedAt = time.Now()
				return report, err
			}
			deployedAgents = item.Scenario.Agents
		}

		result, err := opts.Deps.Engine.Run(ctx, engine.RunInput{
			Cluster:          cl,
			Scenario:         item.Scenario,
			Manager:          opts.Deps.Manager,
			ScenarioPath:     item.Path,
			SandboxNamespace: opts.SandboxNamespace,
		})
		if err != nil {
			report.Results = append(report.Results, result)
			report.InfraFailed = true
			report.EndedAt = time.Now()
			return report, err
		}

		if !result.Passed && opts.Deps.Collector != nil && result.Failure != nil {
			failure := *result.Failure
			failure.SandboxNamespace = opts.SandboxNamespace
			failure.ArtifactsDir = opts.ArtifactsDir
			artifacts, collectErr := opts.Deps.Collector.Collect(ctx, failure)
			if collectErr != nil {
				report.DiagnosticsErrors[item.Scenario.Name] = collectErr
			} else {
				report.Diagnostics[item.Scenario.Name] = artifacts
			}
		}

		report.Results = append(report.Results, result)
		if opts.FailFast && !result.Passed {
			break
		}
	}

	report.EndedAt = time.Now()
	return report, nil
}

func agentSetsEqual(a, b scenario.AgentSet) bool {
	return reflect.DeepEqual([]string(a), []string(b))
}
