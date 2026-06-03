# Core framework features

kube-agents-test is a high-level testing framework for the kube-agents platform: multiple autonomous agents operating on a shared Kubernetes cluster. Unit tests cover individual agents; this framework targets **system** behavior—agents reacting to the same cluster state, possibly conflicting, racing, or depending on each other's outputs—and verifies that they drive the cluster toward the correct state together.

---

## Problem the framework addresses

Testing one agent in isolation is straightforward. The difficult case is the **multi-agent system**: shared cluster state, interactions, conflicts, races, and dependencies between agent outputs. High-level tests must assert that the agent set, collectively, converges the cluster to the intended outcome.

---

## Core concepts

### Test Scenario

A **test scenario** is a declarative description of a single end-to-end test. It specifies:

1. **Initial cluster state** — Resources to pre-create before the test runs (for example via manifests in `setup`).
2. **Optional trigger** — An event that starts or perturbs the behavior under test: resource mutation, agent restart, or fault injection.
3. **Expected final state** — What the cluster should look like when agents have finished: resource conditions, field values, or absence of resources.
4. **Timeout for convergence** — How long the framework waits for the cluster to reach the expected state before failing the scenario.

Scenarios are expressed as data (YAML), not as test code, so the same definitions can run against different clusters and environments.

### Agent Set

An **agent set** is the list of agents that participate in a scenario. Tests can:

- Run a **subset** of agents to isolate interactions between specific agents.
- Run the **full** agent set for true integration coverage.

The `agents` field in a scenario file names which agents are in scope for that run.

### State Assertion

A **state assertion** is how the framework checks success. It is **not** a one-time snapshot of the API server.

Agents are **eventually consistent**: they observe, reconcile, and may retry over time. Assertions therefore use **polling or watch-based** checks that wait until the expected state holds, or until the scenario timeout expires. This matches how operators and controllers behave in production.

---

## Architecture

The framework is organized around a **test runner** that orchestrates lifecycle and collects results, and three cooperating subsystems:

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

### Test Runner

Orchestrates the overall test lifecycle and aggregates pass/fail and diagnostics. It drives the cluster provider, agent manager, and scenario engine in sequence for each scenario.

### Cluster Provider

Creates and tears down **ephemeral clusters** for tests.

- Supports **kind** for CI (no long-lived infrastructure).
- Supports an **existing kubeconfig** for development or staging against a real cluster.

The framework does **not** own the cluster implementation; it only needs a working **kubeconfig** pointing at the cluster under test. Other backends (for example k3d or a pre-provisioned cluster) fit the same contract as long as kubeconfig access is available.

### Agent Manager

Deploys, restarts, and kills agents **within** the test cluster. Agents can be deployed as pods (production-like) or run as local processes (faster iteration).

Exposed controls include:

- **Start/stop individual agents mid-scenario** — For example to test restart behavior, leader election, or recovery after crash.
- **Inject resource limits or network policies** — To simulate degraded runtime conditions without changing agent code.

### Scenario Engine

Executes one test scenario end to end:

1. **Apply initial state** — From Kubernetes manifests or programmatic resource creation.
2. **Fire the trigger** — Patch, agent action, or fault hook as defined in the scenario.
3. **Poll the cluster** — Until expected state is satisfied or the timeout elapses.
4. **Record outcome** — Pass or fail; on failure, trigger collection of diagnostics (see [Failure diagnostics](#failure-diagnostics)).

---

## Scenario definition (YAML model)

Scenarios are **YAML files**. The design includes a representative structure:

| Section | Role |
|---------|------|
| `name` | Unique scenario identifier |
| `description` | Human-readable intent |
| `agents` | Agent set for this run |
| `setup` | Initial cluster state (for example `manifests` listing fixture paths) |
| `trigger` | Optional perturbation (for example a `patch` to a Deployment) |
| `expect` | List of expected resource state checks |
| `timeout` | Maximum wait for convergence (for example `120s`) |

### Example (from design)

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

In this scenario, two agents interact: a scaling agent desires more replicas while a quota agent enforces a cap; the assertion checks both desired replica count and ready replicas on the Deployment.

Expect blocks can reference resources by API version, kind, name, and namespace, and assert on JSON paths (for example `.spec.replicas`) with expected values.

---

## What this framework tests (and does not)

### In scope

- **Agent-to-agent interaction** — Coordination, conflict resolution, ordering of effects on shared resources.
- **Convergence** — Under normal and degraded conditions, the cluster reaches the declared expected state within the timeout.
- **Recovery** — After agent restart or crash, the system returns to correct behavior.
- **Correct final cluster state** — Resources match expectations (fields, conditions, or absence).

### Out of scope

- **Agent internal logic** — Use unit tests for single-agent behavior.
- **Performance and load** — Treated as a separate concern; not the focus of scenario-based convergence tests.
- **Kubernetes correctness** — The framework assumes the API server and controllers behave as documented; it tests **agent behavior on top of** Kubernetes, not Kubernetes itself.

---

## Failure diagnostics

When a scenario **fails** (timeout or assertion mismatch), the framework collects material to debug **why** the cluster did not converge, without requiring a manual reproduction first:

| Artifact | Purpose |
|----------|---------|
| **Agent logs** | Filtered to the scenario's namespace and relevant resources |
| **Kubernetes events** | In the test namespace, showing scheduling, failures, and warnings |
| **Expected vs actual diff** | Between declared expect blocks and live resource state |
| **Mutation timeline** | Resource changes recorded from a watch stream during the test |

Together, these artifacts explain ordering, errors, and drift from the expected end state.

---

## Fault injection

Scenarios can compose **optional fault hooks** to stress recovery, partitioning, and consistency. The design defines:

| Fault | Mechanism | Purpose |
|-------|-----------|---------|
| Kill agent | Delete pod / kill process | Test recovery and leader re-election |
| Network partition | NetworkPolicy between agent and API server | Test behavior when the agent cannot reach the cluster |
| Slow API server | Inject latency via proxy | Test timeout and retry logic |
| Stale cache | Restart informer without full resync | Test correctness with partial state |
| Resource conflict | Concurrent update from test harness | Test conflict retry logic |

Optional fault hooks can be **composed into scenarios**; a trigger may be fault injection (alongside resource mutation or agent restart).

---

## Implementation plan

The design orders delivery as:

1. **Cluster provider** with `kind` support
2. **Agent manager** deploying agents from container images
3. **Scenario engine** — YAML parsing, state application, polling assertions
4. **Failure diagnostics** collection on failed runs
5. **CLI** — Run scenarios (for example `kube-agents-test run scenarios/`)
6. **CI integration** — GitHub Actions workflow

This document describes intended features; only the README and this docs tree exist in the repository until those components are implemented.

---

## Technology choices

| Choice | Rationale |
|--------|-----------|
| **Go** | Same language as the agents; shared **client-go** usage without an FFI boundary |
| **client-go** | Direct Kubernetes API access, watches, and dynamic client for arbitrary resources |
| **kind** | Ephemeral clusters in CI without external infrastructure |
| **No test-framework coupling** | Scenarios are **data**, not `go test` cases. The runner is a **standalone binary**, not a test suite plugin. That avoids `go test` semantics and allows running the same scenarios against any cluster with a kubeconfig |

---

## Summary

| Area | Feature |
|------|---------|
| Concepts | Test Scenario, Agent Set, State Assertion (eventual consistency) |
| Components | Test Runner, Cluster Provider, Agent Manager, Scenario Engine |
| Input | YAML scenarios with setup, trigger, expect, timeout |
| Scope | Multi-agent convergence and recovery; not unit tests, load tests, or K8s validation |
| Operations | Failure diagnostics bundle; composable fault injection |
| Stack | Go, client-go, kind; data-driven scenarios, standalone CLI |

For the canonical design wording, see [README.md](../README.md).
