package engine

import "errors"

// ErrAssertionTimeout is returned when expect conditions do not converge in time.
var ErrAssertionTimeout = errors.New("assertion timeout")

// ErrAssertionMismatch is returned when a terminal assertion mismatch is detected.
var ErrAssertionMismatch = errors.New("assertion mismatch")
