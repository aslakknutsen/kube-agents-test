package agent

import "errors"

// ErrAgentNotFound is returned when the registry cannot resolve an agent ID.
var ErrAgentNotFound = errors.New("agent not found")
