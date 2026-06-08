package scenario

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var scenarioNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidationContext carries runner-level constraints for scenario validation.
type ValidationContext struct {
	SandboxNamespace   string
	AllowClusterScoped bool
	AllowFaults        bool
}

// Validate checks required fields and internal consistency.
func (s *Scenario) Validate() error {
	if s == nil {
		return fmt.Errorf("scenario is nil")
	}
	if s.Name == "" {
		return fmt.Errorf("scenario name is required")
	}
	if !scenarioNamePattern.MatchString(s.Name) {
		return fmt.Errorf("scenario name %q contains invalid characters", s.Name)
	}
	if len(s.Agents) == 0 {
		return fmt.Errorf("scenario %q: agents must be non-empty", s.Name)
	}
	for i, agent := range s.Agents {
		if strings.TrimSpace(agent) == "" {
			return fmt.Errorf("scenario %q: agents[%d] is empty", s.Name, i)
		}
	}
	if len(s.Setup.Manifests) == 0 {
		return fmt.Errorf("scenario %q: setup.manifests must be non-empty", s.Name)
	}
	if len(s.Expect.Assertions) == 0 {
		return fmt.Errorf("scenario %q: expect must contain at least one assertion", s.Name)
	}
	if s.Expect.Timeout <= 0 {
		return fmt.Errorf("scenario %q: expect.timeout must be greater than zero", s.Name)
	}
	for i, assertion := range s.Expect.Assertions {
		if err := assertion.Validate(); err != nil {
			return fmt.Errorf("scenario %q: expect[%d]: %w", s.Name, i, err)
		}
	}
	if s.Trigger != nil {
		if err := s.Trigger.Validate(); err != nil {
			return fmt.Errorf("scenario %q: %w", s.Name, err)
		}
	}
	return nil
}

// ValidateWith applies runner-level sandbox and fault constraints.
func (s *Scenario) ValidateWith(ctx ValidationContext) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if s.Trigger != nil && s.Trigger.Fault != nil && !ctx.AllowFaults {
		return fmt.Errorf("scenario %q: fault triggers require AllowFaults", s.Name)
	}
	if ctx.SandboxNamespace != "" {
		if s.Trigger != nil && s.Trigger.Patch != nil {
			if err := validateNamespacedResource(s.Trigger.Patch.Namespace, ctx); err != nil {
				return fmt.Errorf("scenario %q: trigger.patch: %w", s.Name, err)
			}
		}
		for i, assertion := range s.Expect.Assertions {
			if err := validateNamespacedResource(assertion.Resource.Namespace, ctx); err != nil {
				return fmt.Errorf("scenario %q: expect[%d]: %w", s.Name, i, err)
			}
		}
	}
	return nil
}

func validateNamespacedResource(namespace string, ctx ValidationContext) error {
	if namespace == "" {
		if !ctx.AllowClusterScoped {
			return fmt.Errorf("cluster-scoped resource not allowed without AllowClusterScoped")
		}
		return nil
	}
	if !ctx.AllowClusterScoped && namespace != ctx.SandboxNamespace {
		return fmt.Errorf("namespace %q is outside sandbox %q", namespace, ctx.SandboxNamespace)
	}
	return nil
}

// Validate validates a resource assertion.
func (a *ResourceAssertion) Validate() error {
	if a.Resource.APIVersion == "" {
		return fmt.Errorf("resource apiVersion is required")
	}
	if a.Resource.Kind == "" {
		return fmt.Errorf("resource kind is required")
	}
	if a.Resource.Name == "" {
		return fmt.Errorf("resource name is required")
	}
	if len(a.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}
	for i, cond := range a.Conditions {
		if cond.Path == "" {
			return fmt.Errorf("conditions[%d].path is required", i)
		}
	}
	return nil
}

// ValidateManifestPaths rejects manifest paths that escape the scenario directory.
func ValidateManifestPaths(scenarioDir string, manifests []string) error {
	root, err := filepath.Abs(scenarioDir)
	if err != nil {
		return fmt.Errorf("scenario directory: %w", err)
	}
	root = filepath.Clean(root) + string(filepath.Separator)

	for _, manifest := range manifests {
		if filepath.IsAbs(manifest) {
			return fmt.Errorf("manifest path %q must be relative", manifest)
		}
		if strings.Contains(manifest, "..") {
			return fmt.Errorf("manifest path %q must not contain ..", manifest)
		}
		resolved := filepath.Clean(filepath.Join(root, manifest))
		if !strings.HasPrefix(resolved+string(filepath.Separator), root) {
			return fmt.Errorf("manifest path %q escapes scenario directory", manifest)
		}
	}
	return nil
}
