package diagnostics

import "errors"

// ErrNotImplemented is returned by stub collector until roadmap step 4 is complete.
var ErrNotImplemented = errors.New("diagnostics: not implemented")
