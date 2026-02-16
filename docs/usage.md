# Framework Usage

## Overview

`kube-agents-test` is a Go testing framework for verifying that multiple autonomous agents drive a shared Kubernetes cluster toward the correct state. Tests are written as YAML scenario files and executed via `go test`.

## Directory Layout

```
pkg/
  scenario/    # Scenario types and YAML loader
  cluster/     # Cluster provider interface (kind, existing)
  agent/       # Agent manager interface (pod-based deployment)
  engine/      # Scenario execution: setup, trigger, poll, assert
  diagnostics/ # Failure diagnostics collection
internal/
  testutil/    # go test integration helpers (Framework, Setup, RunScenario)
test/
  scenarios/   # YAML scenario files
  fixtures/    # Kubernetes manifests used by scenarios
  integration_test.go  # Entry point for go test
```

## Writing a Scenario

A scenario is a YAML file that describes:

1. Which agents participate
2. Initial cluster state (manifests to apply)
3. A trigger (resource mutation)
4. Expected final state (field-level conditions on resources)
5. A timeout for convergence

Example (`test/scenarios/scaling-respects-quota.yaml`):

```yaml
name: scaling-agent-respects-quota-agent
description: >
  When the scaling agent wants to add replicas but the quota agent has
  capped the namespace, the deployment should stay at the capped count.

agents:
  - scaling-agent
  - quota-agent

setup:
  manifests:
    - fixtures/namespace-with-quota.yaml
    - fixtures/deployment-at-limit.yaml

trigger:
  patch:
    apiVersion: apps/v1
    kind: Deployment
    name: target
    namespace: test
    spec:
      replicas: 10

expect:
  - resource:
      apiVersion: apps/v1
      kind: Deployment
      name: target
      namespace: test
    conditions:
      - path: .spec.replicas
        value: 5
      - path: .status.readyReplicas
        value: 5
timeout: 120s
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Unique scenario identifier, used as the `go test` subtest name |
| `description` | no | Human-readable explanation |
| `agents` | yes | List of agent names to deploy |
| `setup.manifests` | no | YAML files to apply before the test starts |
| `trigger.patch` | no | Strategic merge patch applied to kick off the test |
| `expect[].resource` | yes | Identifies the resource to check (apiVersion, kind, name, namespace) |
| `expect[].conditions` | yes | List of `{path, value}` pairs; path is a dot-separated JSONPath |
| `timeout` | no | How long to poll before failing (default: 2m) |

## Running Tests

### Validate scenario loading only (no cluster needed)

```bash
go test -short -v ./test/
```

This parses all YAML scenarios and checks for structural errors. Fast, no dependencies.

### Run against a kind cluster

```bash
go test -v ./test/ -timeout 10m
```

The framework creates a kind cluster, deploys agents as pods, runs all scenarios, and tears down the cluster. Requires `kind` and `docker` on the PATH.

### Run against an existing cluster

```bash
KUBECONFIG=~/.kube/config go test -v ./test/ -timeout 10m
```

Skips cluster creation/teardown. Useful during development when you already have a cluster running.

### Run a single scenario

```bash
go test -v -run TestScalingRespectsQuota ./test/ -timeout 10m
```

Uses the standard `go test -run` flag to filter by test name.

## Writing Go Tests

The `testutil.Framework` type is the main entry point. It wires together the cluster provider, agent manager, and scenario engine.

### Minimal example

```go
package test

import (
    "testing"

    "github.com/kube-agents/kube-agents-test/internal/testutil"
    "github.com/kube-agents/kube-agents-test/pkg/agent"
)

func TestMyScenario(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    f := testutil.Setup(t, testutil.Options{
        AgentConfigs: []agent.AgentConfig{
            {Name: "my-agent", Image: "my-registry/my-agent:latest", Namespace: "agents"},
        },
    })

    f.RunScenario(t, "scenarios/my-scenario.yaml")
}
```

### Run all scenarios in a directory

```go
func TestAllScenarios(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    f := testutil.Setup(t, testutil.Options{...})
    f.RunScenarioDir(t, "scenarios")
}
```

Each scenario file becomes a `t.Run` subtest, so you get per-scenario pass/fail in the output.

### Custom cluster provider

```go
f := testutil.Setup(t, testutil.Options{
    ClusterProvider: &cluster.ExistingProvider{
        KubeconfigPath: "/path/to/kubeconfig",
    },
})
```

## Cluster Providers

| Provider | When to use |
|----------|-------------|
| `KindProvider` | CI, ephemeral clusters, default |
| `ExistingProvider` | Dev/staging, already have a cluster |

Implement the `cluster.Provider` interface to add support for other providers (k3d, EKS, etc.).

## Agent Manager

The `PodManager` deploys agents as Kubernetes pods. Configure agents with `AgentConfig`:

```go
agent.AgentConfig{
    Name:      "scaling-agent",
    Image:     "ghcr.io/kube-agents/scaling-agent:latest",
    Namespace: "kube-agents",
    Args:      []string{"--log-level=debug"},
}
```

Implement the `agent.Manager` interface for alternative deployment strategies (local processes, Helm charts, etc.).

## Failure Diagnostics

When a scenario fails, the framework can collect:

- Agent logs (filtered to the scenario's namespace)
- Kubernetes events
- Resource diffs (expected vs actual field values)

Pass a `diagnostics.Collector` implementation via `testutil.Options` to enable this. The diagnostics appear in the `go test` output under the failing subtest.

## Implementation Status

The core framework components are implemented:

| Component | Status | Details |
|-----------|--------|---------|
| `PodManager.Deploy` / `Stop` / `Logs` | Done | Creates Kubernetes Deployments via client-go, collects pod logs |
| `Engine.applyManifest` | Done | Reads multi-document YAML, creates/updates resources via dynamic client |
| `Engine.applyPatch` | Done | Applies merge patches to existing resources |
| `Engine.checkExpectation` | Done | Fetches resources, evaluates dot-path conditions with numeric type coercion |
| `ClusterCollector` | Done | Collects agent pod logs and namespace events on failure |
| Fault injection hooks | Not yet | Network partition, kill agent, slow API server, etc. |

### Test coverage

- `pkg/scenario/` — 4 tests (loader, validation, directory loading)
- `pkg/engine/` — 17 tests (path lookup, value comparison)
- `pkg/agent/` — 5 tests (deploy, stop, stop-all, unknown agent, default namespace)
- `test/` — scenario YAML parsing (short mode), integration tests (require cluster)

## Extending the Framework

### Adding a new cluster provider

Implement the `cluster.Provider` interface:

```go
type Provider interface {
    Create(ctx context.Context) (kubeconfigPath string, err error)
    Destroy(ctx context.Context) error
    Kubeconfig() (string, error)
}
```

### Adding a new agent deployment strategy

Implement the `agent.Manager` interface:

```go
type Manager interface {
    Deploy(ctx context.Context, agentName string) error
    Stop(ctx context.Context, agentName string) error
    StopAll(ctx context.Context) error
    Logs(ctx context.Context, agentName string) (string, error)
}
```

For example, a `LocalProcessManager` could run agents as local binaries instead of pods.

### Adding a custom diagnostics collector

Implement `diagnostics.Collector`:

```go
type Collector interface {
    Collect(ctx context.Context, scope Scope) (*Report, error)
}
```

### Adding new trigger types

Extend the `scenario.Trigger` struct and the engine's `applyTrigger` method. For example, to support creating a resource as a trigger (not just patching), add a `Create` field to the trigger type and handle it in the engine.
