package agent

import "errors"

var (
	// ErrInvalidDeployMode is returned when agent mode string is not recognized.
	ErrInvalidDeployMode = errors.New("agent: invalid deploy mode")
	// ErrNotImplemented is returned by stub manager until roadmap step 2 is complete.
	ErrNotImplemented = errors.New("agent: not implemented")
)
