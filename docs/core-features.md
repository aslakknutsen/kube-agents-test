# Core framework features

This document elaborates the kube-agents-test framework as described in the project [README](../README.md). It is intended for later inclusion in the overall user documentation set.

## Problem the framework solves

Unit testing individual agents is straightforward. The difficult part is testing the **system**: multiple agents reacting to the same cluster state, potentially conflicting, racing, or depending on each other's outputs. High-level tests must verify that the agents, taken together, drive the cluster toward the correct state.

## Core concepts

### Test scenario

A **test scenario** is a declarative description of:

1. **Initial cluster state** — resources to pre-create before the test runs
2. **Optional trigger** — a resource mutation, agent restart, or fault injection that starts or perturbs the behavior under test
3. **Expected final state** — resource conditions, fields, or absence that define success
4. **Timeout for convergence** — how long the framework waits for the cluster to reach the expected state

Scenarios are data (YAML), not test code. That keeps scenarios portable and decoupled from any single test-runner semantics.

### Agent set

An **agent set** specifies which agents participate in a scenario. Tests can run:

- A **subset** of agents — to isolate interactions between specific agents
- The **full set** — for true integration tests across the platform

This lets the same scenario machinery support focused and end-to-end runs.

### State assertion

A **state assertion** is a polling- or watch-based check that waits for the cluster to converge to the expected state within the timeout. It is **not** a point-in-time snapshot: agents are eventually consistent, so assertions must be eventually consistent as well.

The scenario engine polls (or watches) until expectations are met or the timeout expires.

## Architecture

The framework is organized around a test runner and three cooperating subsystems:

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

### Test runner

The **test runner** orchestrates the lifecycle of a run and collects results (pass/fail and diagnostics). It coordinates the cluster provider, agent manager, and scenario engine.

### Cluster provider

The **cluster provider** creates and tears down ephemeral clusters. It supports:

- **kind** — for CI, without relying on long-lived infrastructure
- **An existing kubeconfig** — for development and staging against a cluster you already have

The framework does not own the cluster implementation; it needs a **kubeconfig** to talk to whatever cluster backs the test. The diagram also references **k3d** and **real cluster** as deployment options alongside kind.

### Agent manager

The **agent manager** deploys, restarts, and kills agents within the test cluster. Agents can run:

- As **pods** — production-like deployment in the cluster
- As **local processes** — faster iteration during development

Controls exposed for scenarios include:

- **Starting or stopping individual agents** mid-scenario — e.g. restart behavior, leader election
- **Injecting resource limits or network policies** — to simulate degraded runtime conditions

### Scenario engine

The **scenario engine** executes a test scenario end to end:

1. Applies **initial state** (Kubernetes manifests or programmatic resource creation)
2. **Fires the trigger** (if defined)
3. **Polls the cluster** until the expected state is reached or the timeout expires
4. Records **pass/fail** and, on failure, collects **diagnostics** (agent logs, resource diffs, events)

## Scenario definition (YAML model)

Scenarios are defined as **YAML files**. The README includes this illustrative structure (field names and nesting follow that example):

| Section | Role |
|---------|------|
| `name` | Scenario identifier |
| `description` | Human-readable intent |
| `agents` | List of agent names in the agent set |
| `setup.manifests` | Paths to manifests that establish initial cluster state |
| `trigger` | Optional action that perturbs the cluster (example uses a `patch` on a Deployment) |
| `expect` | List of expected resource states (apiVersion, kind, name, namespace, `conditions` on JSON paths) |
| `timeout` | Maximum wait for convergence (example: `120s`). In the design example below, `timeout` is nested under `expect` alongside the last resource entry; the README does not specify whether `timeout` is a top-level field or nested under `expect` |

### Example scenario (from design)

The design README documents a scenario named `scaling-agent-respects-quota-agent`: two agents (`scaling-agent`, `quota-agent`), setup manifests for a namespace with quota and a deployment at the limit, a trigger that patches replica count to 10, and expectations that replicas and ready replicas remain at 5 because the quota agent caps the deployment.

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

Path-based `conditions` (e.g. `.spec.replicas`) express expected field values on watched resources. The exact schema for triggers beyond `patch` and for additional expect shapes will be defined when the scenario engine is implemented; this document only records what the design README shows.

## What the framework tests (and does not)

### In scope

- **Agent-to-agent interaction** — coordination, conflict resolution, ordering
- **Convergence** under normal and degraded conditions
- **Recovery** after agent restart or crash
- **Correct final state** of cluster resources

### Out of scope

- **Agent internal logic** — use unit tests
- **Performance and load** — treated as a separate concern
- **Kubernetes correctness** — the cluster API and control plane are assumed correct; tests target agent behavior on top of Kubernetes

## Failure diagnostics

When a scenario **fails**, the framework collects material to explain why the cluster did not converge, without requiring a manual reproduction first:

| Artifact | Purpose |
|----------|---------|
| **Agent logs** | Filtered to the scenario's namespace and resources |
| **Kubernetes events** | In the test namespace |
| **Resource diff** | Between expected and actual state |
| **Mutation timeline** | From a watch stream recorded during the test |

This gives enough to debug why the cluster did not converge without having to reproduce the failure manually.

## Fault injection

Optional **fault hooks** can be composed into scenarios. The design defines these hooks:

| Fault | Mechanism | Purpose |
|-------|-----------|---------|
| Kill agent | Delete pod / kill process | Test recovery and leader re-election |
| Network partition | NetworkPolicy between agent and API server | Test agent behavior when it cannot reach the cluster |
| Slow API server | Inject latency via proxy | Test timeout and retry logic |
| Stale cache | Restart informer without full resync | Test agent correctness with partial state |
| Resource conflict | Concurrent update from test harness | Test conflict retry logic |

Faults are optional; a scenario may run with setup, trigger, and expect only. The design README assigns each fault a mechanism in the table above; it does not specify which component applies each mechanism.

## Planned implementation (roadmap)

The README lists this implementation order (not yet built in the repository):

1. Cluster provider with `kind` support
2. Agent manager that deploys agents from container images
3. Scenario engine: YAML parsing, state application, polling assertions
4. Failure diagnostics collection
5. CLI to run scenarios (`kube-agents-test run scenarios/`)
6. CI integration (GitHub Actions workflow)

## Technology choices

| Choice | Rationale |
|--------|-----------|
| **Go** | Same language as the agents; shared `client-go` usage; no FFI boundary |
| **client-go** | Direct Kubernetes API interaction, watches, dynamic client for arbitrary resources |
| **kind** | Ephemeral clusters in CI without infrastructure dependencies |
| **No test framework dependency** | Scenarios are data, not code; the runner is a standalone binary, not a `go test` suite — avoids coupling to Go test semantics and allows running scenarios against any cluster |

## Relationship to other documentation

- [docs/README.md](README.md) — index for this documentation folder
- [README.md](../README.md) — original design document and source of truth until implementation docs exist
