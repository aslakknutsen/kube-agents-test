package errs

import "errors"

// ErrNotImplemented indicates a subsystem backend is not yet implemented.
var ErrNotImplemented = errors.New("not implemented")
