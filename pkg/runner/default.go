package runner

import (
	"context"
	"errors"
	"fmt"

	intcluster "github.com/kube-agents/kube-agents-test/internal/cluster"
	agentstub "github.com/kube-agents/kube-agents-test/internal/agent/stub"
	diagstub "github.com/kube-agents/kube-agents-test/internal/diagnostics/stub"
	scenariostub "github.com/kube-agents/kube-agents-test/internal/scenario/stub"
	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// defaultRunner orchestrates subsystem lifecycle for each scenario.
type defaultRunner struct {
	cfg Config
}

// New builds a Runner, filling nil dependencies with internal stubs.
func New(cfg Config) (Runner, error) {
	if cfg.ClusterProvider == nil {
		p, err := intcluster.NewProvider(cfg.EphemeralOpts)
		if err != nil {
			return nil, err
		}
		cfg.ClusterProvider = p
	}
	if cfg.ScenarioEngine == nil {
		cfg.ScenarioEngine = scenariostub.New()
	}
	if cfg.Diagnostics == nil {
		cfg.Diagnostics = diagstub.New()
	}
	if cfg.AgentManagerFactory == nil {
		cfg.AgentManagerFactory = func(c *cluster.Cluster, ac agent.Config) (agent.Manager, error) {
			return agentstub.New(c, ac), nil
		}
	}
	return &defaultRunner{cfg: cfg}, nil
}

// Run executes scenarios sequentially: cluster → agents → engine → diagnostics on fail → teardown.
func (r *defaultRunner) Run(ctx context.Context, scenarios []*scenario.Scenario) (*SuiteResult, error) {
	if len(scenarios) == 0 {
		return &SuiteResult{}, nil
	}

	cl, err := r.ensureCluster(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r.cfg.ClusterMode == cluster.ModeEphemeral {
			_ = r.cfg.ClusterProvider.Teardown(ctx)
		}
	}()

	mgr, err := r.ensureAgentManager(cl)
	if err != nil {
		return nil, err
	}

	suite := &SuiteResult{
		Results: make([]ScenarioRunResult, 0, len(scenarios)),
	}
	for _, sc := range scenarios {
		suite.Results = append(suite.Results, r.runOne(ctx, cl, mgr, sc))
	}
	for _, res := range suite.Results {
		if res.Passed {
			suite.Passed++
		} else {
			suite.Failed++
		}
	}
	return suite, nil
}

func (r *defaultRunner) ensureCluster(ctx context.Context) (*cluster.Cluster, error) {
	switch r.cfg.ClusterMode {
	case cluster.ModeAttached:
		if r.cfg.Kubeconfig == "" {
			return nil, fmt.Errorf("runner: kubeconfig required for attached mode")
		}
		return r.cfg.ClusterProvider.Attach(ctx, r.cfg.Kubeconfig)
	case cluster.ModeEphemeral:
		return r.cfg.ClusterProvider.Provision(ctx)
	default:
		return nil, fmt.Errorf("runner: unknown cluster mode %v", r.cfg.ClusterMode)
	}
}

func (r *defaultRunner) ensureAgentManager(cl *cluster.Cluster) (agent.Manager, error) {
	if r.cfg.AgentManager != nil {
		return r.cfg.AgentManager, nil
	}
	return r.cfg.AgentManagerFactory(cl, r.cfg.AgentConfig)
}

func (r *defaultRunner) runOne(ctx context.Context, cl *cluster.Cluster, mgr agent.Manager, sc *scenario.Scenario) ScenarioRunResult {
	out := ScenarioRunResult{Name: sc.Name}

	if err := mgr.DeploySet(ctx, sc.Agents); err != nil {
		out.Err = fmt.Errorf("deploy agents: %w", err)
		_ = mgr.Teardown(ctx)
		return out
	}
	defer func() { _ = mgr.Teardown(ctx) }()

	result, err := r.cfg.ScenarioEngine.Run(ctx, sc, scenario.RunOptions{
		Cluster: cl,
		Agents:  mgr,
	})
	if err != nil {
		out.Err = fmt.Errorf("run scenario: %w", err)
		return out
	}
	out.Result = result
	out.Passed = result != nil && result.Passed

	if !out.Passed && r.cfg.Diagnostics != nil && result != nil && result.Failure != nil {
		bundle, derr := r.cfg.Diagnostics.Collect(ctx, diagnostics.CollectRequest{
			Cluster:  cl,
			Scenario: sc,
			Failure:  result.Failure,
		})
		if derr != nil && !errors.Is(derr, diagnostics.ErrNotImplemented) {
			out.Err = errors.Join(out.Err, fmt.Errorf("collect diagnostics: %w", derr))
		} else if bundle != nil {
			out.Diagnostics = bundle
		}
	}
	return out
}
