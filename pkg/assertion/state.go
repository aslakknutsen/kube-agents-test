package assertion

import (
	"context"
	"fmt"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Asserter checks that the cluster has converged to the expected state.
type Asserter interface {
	// WaitForState polls the cluster until all resource expectations are met
	// or the context deadline expires.
	WaitForState(ctx context.Context, expectations []scenario.ResourceExpectation) error
}

// PollingAsserter implements Asserter by polling at a fixed interval.
type PollingAsserter struct {
	Kubeconfig   string
	PollInterval time.Duration
}

func NewPollingAsserter(kubeconfig string, interval time.Duration) *PollingAsserter {
	return &PollingAsserter{
		Kubeconfig:   kubeconfig,
		PollInterval: interval,
	}
}

func (p *PollingAsserter) WaitForState(ctx context.Context, expectations []scenario.ResourceExpectation) error {
	ticker := time.NewTicker(p.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for expected state: %w", ctx.Err())
		case <-ticker.C:
			met, err := p.checkAll(ctx, expectations)
			if err != nil {
				return fmt.Errorf("checking expectations: %w", err)
			}
			if met {
				return nil
			}
		}
	}
}

func (p *PollingAsserter) checkAll(ctx context.Context, expectations []scenario.ResourceExpectation) (bool, error) {
	for _, exp := range expectations {
		met, err := p.checkOne(ctx, exp)
		if err != nil {
			return false, err
		}
		if !met {
			return false, nil
		}
	}
	return true, nil
}

func (p *PollingAsserter) checkOne(ctx context.Context, exp scenario.ResourceExpectation) (bool, error) {
	// TODO: use client-go dynamic client to fetch the resource and compare
	// each condition's path/value against the actual object.
	_ = exp
	return true, nil
}
