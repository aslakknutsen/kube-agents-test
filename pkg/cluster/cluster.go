package cluster

import "k8s.io/client-go/rest"

// Cluster is an opaque handle to a Kubernetes cluster used by agents and the engine.
type Cluster interface {
	ID() string
	RESTConfig() (*rest.Config, error)
	KubeconfigPath() (string, bool)
}
