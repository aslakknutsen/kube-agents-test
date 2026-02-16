# Usage

This doc walks through writing and running a test scenario against kube-agents.

## Prerequisites

- Go 1.23+
- [kind](https://kind.sigs.k8s.io/) (if you want ephemeral clusters)
- A set of agent container images or local binaries to test

## Writing a Scenario

A scenario is a YAML file that describes the initial cluster state, a trigger, and the expected outcome.

### Minimal example

Create `scenarios/my-agent-recovers-from-restart.yaml`:

```yaml
name: my-agent-recovers-from-restart
description: >
  After the reconciler agent is killed and restarted, it should
  re-reconcile the pending ConfigMap within the timeout.

agents:
  - reconciler-agent

setup:
  manifests:
    - fixtures/namespace.yaml
    - fixtures/pending-configmap.yaml

expect:
  resources:
    - resource:
        apiVersion: v1
        kind: ConfigMap
        name: pending
        namespace: test
      conditions:
        - path: .data.status
          value: reconciled
  timeout: 60s
```

No `trigger` block here — the scenario just checks that the agent converges on its own after setup.

### With a trigger

If you need to mutate a resource mid-test to provoke agent behavior:

```yaml
trigger:
  patch:
    apiVersion: apps/v1
    kind: Deployment
    name: target
    namespace: test
    spec:
      replicas: 10
```

The engine applies this patch after deploying the agents, then starts polling for the expected state.

### Multiple expectations

You can assert on more than one resource:

```yaml
expect:
  resources:
    - resource:
        apiVersion: apps/v1
        kind: Deployment
        name: web
        namespace: test
      conditions:
        - path: .spec.replicas
          value: 3
    - resource:
        apiVersion: v1
        kind: ConfigMap
        name: web-config
        namespace: test
      conditions:
        - path: .data.version
          value: "2"
  timeout: 90s
```

All expectations must be satisfied before the timeout.

## Writing a Go Test

The framework plugs into `go test`. A typical test file looks like this:

```go
package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kube-agents/kube-agents-test/pkg/cluster"
	"github.com/kube-agents/kube-agents-test/pkg/framework"
)

var fw *framework.Framework

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Use an existing cluster if KUBECONFIG is set, otherwise create a kind cluster.
	var provider cluster.Provider
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		provider = cluster.NewExistingClusterProvider(kc)
	} else {
		provider = cluster.NewKindProvider("my-test-cluster")
	}

	var err error
	fw, err = framework.New(ctx, framework.Config{
		ClusterProvider: provider,
	})
	if err != nil {
		panic("framework setup failed: " + err.Error())
	}

	code := m.Run()
	_ = fw.Teardown(ctx)
	os.Exit(code)
}

func TestReconcilerRecoversFromRestart(t *testing.T) {
	fw.RunScenario(t, filepath.Join("..", "scenarios", "my-agent-recovers-from-restart.yaml"))
}
```

`TestMain` sets up the cluster once for all tests in the package. Each `Test*` function points at a scenario YAML and the framework handles the rest: apply setup manifests, deploy agents, fire the trigger (if any), poll for expected state, and collect diagnostics on failure.

## Running Tests

Against a kind cluster (created automatically):

```bash
go test ./test/ -v -timeout 5m
```

Against an existing cluster:

```bash
KUBECONFIG=~/.kube/config go test ./test/ -v -timeout 5m
```

### Running a single scenario

```bash
go test ./test/ -v -run TestReconcilerRecoversFromRestart -timeout 5m
```

## What Happens on Failure

When a scenario doesn't converge within the timeout, the framework collects:

- Agent pod logs (filtered to the test namespace)
- Kubernetes events
- Diff between expected and actual resource state

These are printed via `t.Log` so they show up in the `go test -v` output.

## Fixture Manifests

Put your Kubernetes YAML fixtures in `scenarios/fixtures/`. These are plain manifests — namespaces, deployments, configmaps, resource quotas, etc. — that the engine applies before deploying agents.

Example `scenarios/fixtures/namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: test
```

Paths in the scenario `setup.manifests` list are relative to the scenario file's directory.
