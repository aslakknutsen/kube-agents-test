package cluster

import "errors"

// ErrClusterUnavailable is returned when cluster ensure or attach fails.
var ErrClusterUnavailable = errors.New("cluster unavailable")
