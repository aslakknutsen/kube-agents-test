package scenario

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// timeDuration wraps time.Duration for YAML unmarshaling (e.g. "120s").
type timeDuration struct {
	Duration time.Duration
}

func (d *timeDuration) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}
