package runner

import "errors"

// ErrDuplicateScenario is returned when a suite contains duplicate scenario names.
var ErrDuplicateScenario = errors.New("duplicate scenario name")
