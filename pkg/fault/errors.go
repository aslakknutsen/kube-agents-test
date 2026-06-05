package fault

import "errors"

// ErrInvalidTrigger is returned when a trigger or fault definition fails validation.
var ErrInvalidTrigger = errors.New("invalid trigger")
