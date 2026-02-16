package assertion

import (
	"context"
	"fmt"
	"time"
)

// Condition represents a single field-level assertion on a Kubernetes resource.
type Condition struct {
	Path  string
	Value interface{}
}

// ResourceChecker describes a resource and the conditions it must satisfy.
type ResourceChecker struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	Conditions []Condition
}

// PollUntilMatch polls the cluster until the resource matches all conditions
// or the context deadline is exceeded. This is the core of eventually-consistent
// assertion: it retries because agents may not have converged yet.
func PollUntilMatch(ctx context.Context, checker ResourceChecker) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return fmt.Errorf("timed out waiting for %s/%s %s/%s: %w",
					checker.APIVersion, checker.Kind, checker.Namespace, checker.Name, lastErr)
			}
			return ctx.Err()
		case <-ticker.C:
			err := checkResource(checker)
			if err == nil {
				return nil
			}
			lastErr = err
		}
	}
}

// checkResource fetches the resource and evaluates all conditions.
func checkResource(checker ResourceChecker) error {
	// TODO: use client-go dynamic client to fetch the resource,
	// then walk the unstructured object using checker.Conditions[].Path
	// and compare against checker.Conditions[].Value.
	//
	// For now, return an error to indicate not implemented.
	return fmt.Errorf("not implemented: resource check for %s/%s %s/%s",
		checker.APIVersion, checker.Kind, checker.Namespace, checker.Name)
}
