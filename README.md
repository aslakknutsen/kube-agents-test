# kube-agents-test

High-level testing framework for the kube-agents platform — a system where multiple autonomous agents operate on a shared Kubernetes cluster.

## Problem

Unit testing individual agents is straightforward. The hard part is testing the *system*: multiple agents reacting to the same cluster state, potentially conflicting, racing, or depending on each other's outputs. High-level tests need to verify that the agents, taken together, drive the cluster toward the correct state.

## Design

### Core Concepts

**Test Scenario** — A declarative description of:
1. Initial cluster state (resources to pre-create)
2. An optional trigger (resource mutation, agent restart, fault injection)
3. Expected final state (resource conditions, fields, or absence)
4. Timeout for convergence

**Agent Set** — Which agents participate in the scenario. Tests can run a subset of agents to isolate interactions or the full set for true integration tests.

**State Assertion** — A polling/watch-based check that waits for the cluster to converge to the expected state within the timeout. Not a point-in-time snapshot — agents are eventually consistent, so assertions must be too.

### Architecture

```
┌──────────────────────────────────────────────────┐
│                  Test Runner                      │
│  (orchestrates lifecycle, collects results)       │
└──────┬──────────────┬─────────────────┬──────────┘
       │              │                 │
       ▼              ▼                 ▼
┌────────────┐ ┌─────────────┐ ┌───────────────┐
│  Cluster   │ │   Agent     │ │   Scenario    │
│  Provider  │ │   Manager   │ │   Engine      │
└────────────┘ └─────────────┘ └───────────────┘
       │              │                 │
       ▼              ▼                 ▼
   kind/k3d/     deploy/start/    apply initial
   real cluster   stop agents     state, inject
                                  faults, assert
```

**Cluster Provider** — Creates and tears down ephemeral clusters. Supports `kind` for CI and an existing kubeconfig for dev/staging. The framework doesn't own the cluster implementation — it just needs a kubeconfig.

**Agent Manager** — Deploys, restarts, and kills agents within the test cluster. Agents can be deployed as pods (production-like) or run as local processes (faster iteration). The manager exposes controls for:
- Starting/stopping individual agents mid-scenario (to test restart behavior, leader election, etc.)
- Injecting resource limits or network policies to simulate degraded conditions

**Scenario Engine** — Executes a test scenario:
1. Applies initial state (Kubernetes manifests or programmatic resource creation)
2. Fires the trigger
3. Polls the cluster until the expected state is reached or the timeout expires
4. Records pass/fail and collects diagnostics on failure (agent logs, resource diffs, events)

### Scenario Definition

Scenarios are YAML files:

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
      replicas: 10  # scaling agent wants 10

expect:
  - resource:
      apiVersion: apps/v1
      kind: Deployment
      name: target
      namespace: test
    conditions:
      - path: .spec.replicas
        value: 5  # quota agent should cap it
      - path: .status.readyReplicas
        value: 5
  timeout: 120s
```

### What This Tests (and Doesn't)

**In scope:**
- Agent-to-agent interaction (coordination, conflict resolution, ordering)
- Convergence under normal and degraded conditions
- Recovery after agent restart or crash
- Correct final state of cluster resources

**Out of scope:**
- Agent internal logic (use unit tests)
- Performance/load (separate concern)
- Kubernetes itself (assumed correct)

### Failure Diagnostics

When a scenario fails, the framework collects:
- Agent logs (filtered to the scenario's namespace/resources)
- Kubernetes events in the test namespace
- Diff between expected and actual resource state
- Timeline of resource mutations (from a watch stream recorded during the test)

This gives enough to debug *why* the cluster didn't converge without having to reproduce the failure manually.

### Fault Injection

Optional fault hooks that can be composed into scenarios:

| Fault | Mechanism | Purpose |
|-------|-----------|---------|
| Kill agent | Delete pod / kill process | Test recovery and leader re-election |
| Network partition | NetworkPolicy between agent and API server | Test agent behavior when it can't reach the cluster |
| Slow API server | Inject latency via proxy | Test timeout and retry logic |
| Stale cache | Restart informer without full resync | Test agent correctness with partial state |
| Resource conflict | Concurrent update from test harness | Test conflict retry logic |

### Implementation Plan

1. Cluster provider with `kind` support
2. Agent manager that deploys agents from container images
3. Scenario engine: YAML parsing, state application, polling assertions
4. Failure diagnostics collection
5. CLI to run scenarios (`kube-agents-test run scenarios/`)
6. CI integration (GitHub Actions workflow)

### Tech Choices

- **Go** — same language as the agents, shared client-go usage, no FFI boundary
- **client-go** — direct Kubernetes API interaction, watches, dynamic client for arbitrary resources
- **kind** — ephemeral clusters in CI without infrastructure dependencies
- **No test framework dependency** — scenarios are data, not code. The runner is a standalone binary, not a test suite. This avoids coupling to `go test` semantics and makes it possible to run scenarios against any cluster.
