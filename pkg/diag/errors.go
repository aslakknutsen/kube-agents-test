package diag

import "errors"

// ErrDiagnosticsPartial is returned when some diagnostic artifacts could not be collected.
var ErrDiagnosticsPartial = errors.New("diagnostics partial")
