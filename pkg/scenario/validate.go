package scenario

import (
	"fmt"
	"strings"
	"time"
)

// Validate checks required fields and basic constraints.
func (s *Scenario) Validate() error {
	if s == nil {
		return fmt.Errorf("scenario is nil")
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(s.Agents) == 0 {
		return fmt.Errorf("agents: at least one agent is required")
	}
	for i, a := range s.Agents {
		if strings.TrimSpace(a) == "" {
			return fmt.Errorf("agents[%d]: empty agent name", i)
		}
	}
	if len(s.Expect.Resources) == 0 {
		return fmt.Errorf("expect: at least one resource assertion is required")
	}
	if s.Expect.Timeout <= 0 {
		return fmt.Errorf("expect.timeout: must be positive, got %s", s.Expect.Timeout)
	}
	for i, re := range s.Expect.Resources {
		if err := re.validate(i); err != nil {
			return err
		}
	}
	if s.Trigger != nil && s.Trigger.Patch != nil {
		if err := s.Trigger.Patch.ObjectRef.validate("trigger.patch"); err != nil {
			return err
		}
	}
	return nil
}

func (re *ResourceExpect) validate(idx int) error {
	prefix := fmt.Sprintf("expect.resources[%d]", idx)
	if err := re.Resource.validate(prefix + ".resource"); err != nil {
		return err
	}
	if len(re.Conditions) == 0 {
		return fmt.Errorf("%s: at least one condition is required", prefix)
	}
	for j, c := range re.Conditions {
		if strings.TrimSpace(c.Path) == "" {
			return fmt.Errorf("%s.conditions[%d]: path is required", prefix, j)
		}
	}
	return nil
}

func (o *ObjectRef) validate(prefix string) error {
	if strings.TrimSpace(o.APIVersion) == "" {
		return fmt.Errorf("%s.apiVersion is required", prefix)
	}
	if strings.TrimSpace(o.Kind) == "" {
		return fmt.Errorf("%s.kind is required", prefix)
	}
	if strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf("%s.name is required", prefix)
	}
	return nil
}

// DefaultTimeout is used when callers need a fallback; scenarios must set expect.timeout.
const DefaultTimeout = 120 * time.Second
