package cluster

import "context"

// Provider provisions and tears down clusters. It never deploys agents.
type Provider interface {
	Ensure(ctx context.Context, cfg Config) (Cluster, error)
	Teardown(ctx context.Context, c Cluster) error
}
