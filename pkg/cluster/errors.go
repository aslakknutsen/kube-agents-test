package cluster

import "errors"

var (
	// ErrInvalidMode is returned when cluster mode string is not recognized.
	ErrInvalidMode = errors.New("cluster: invalid mode")
	// ErrNotImplemented is returned by stub backends until roadmap step 1 is complete.
	ErrNotImplemented = errors.New("cluster: not implemented")
)
