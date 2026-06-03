# Core Features

This page defines the core features of kube-agents-test: the concepts you use to describe a test, the components that execute it, and the capabilities around failure analysis and fault injection.

All content here is derived from the project design. Where behavior is not yet specified in implementation, gaps are called out explicitly.

## Core concepts

Three concepts form the basis of every test: **Test Scenario**, **Agent Set**, and **State Assertion**.

### Test Scenario

A **Test Scenario** is a declarative description of a multi-agent integration test. It specifies the full lifecycle of a single test run:

1. **Initial cluster state** — Resources to pre-create before agents act (for example, namespaces, quotas, deployments).
2. **Optional trigger** — An event that starts or changes the conditions under test. Triggers can include resource mutations, agent restarts, or fault injection.
3. **Expected final state** — What the cluster should look like after agents converge. This may include resource conditions, specific field values, or the absence of resources.
4. **Timeout for convergence** — How long the framework waits for the cluster to reach the expected state before declaring failure.

Scenarios are **data, not code**. They are defined as YAML files rather than as Go test functions. This keeps tests portable across clusters and decouples scenario definition from the `go test` runner semantics.

**Example use:** A scenario might pre-create a deployment at its replica quota limit, trigger a scale-up request, and assert that a quota agent caps the deployment while a scaling agent respects that cap.

### Agent Set

An **Agent Set** declares which agents participate in a scenario. The `agents` field in a scenario YAML file lists agent names (for example, `scaling-agent`, `quota-agent`).

Tests can run:

- **A subset of agents** — To isolate interactions between specific agents or to test behavior when certain agents are absent.
- **The full agent set** — For true end-to-end integration tests where all agents may interact on shared cluster state.

The Agent Manager (see [Architecture](#architecture)) is responsible for deploying, starting, stopping, and killing agents according to this set. *Implementation detail not yet specified:* how agent names map to container images, deployments, or local processes.

### State Assertion

A **State Assertion** verifies that the cluster has converged to the expected final state. Assertions are **polling- or watch-based**, not point-in-time snapshots.

This matters because agents are **eventually consistent**. A single read immediately after a trigger may show intermediate or stale state. The framework polls (or watches) the cluster until:

- All expected conditions are satisfied, or
- The scenario timeout expires.

Each assertion in the `expect` block targets a Kubernetes resource and checks conditions such as field values at JSON paths (for example, `.spec.replicas`, `.status.readyReplicas`).

**Gap:** The exact polling interval, watch reconnect behavior, and partial-match semantics for complex conditions are not yet defined in the design.

---

## Architecture

The framework is organized around four cooperating components plus a top-level orchestrator.

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

The **Test Runner** orchestrates the full lifecycle of a test run and collects results. It coordinates the Cluster Provider, Agent Manager, and Scenario Engine.

Responsibilities implied by the design:

- Start and tear down the test environment.
- Invoke scenario execution.
- Aggregate pass/fail outcome and failure diagnostics.

**Gap:** CLI flags, exit codes, parallel scenario execution, and result reporting formats are not yet specified.

### Cluster Provider

The **Cluster Provider** creates and tears down ephemeral Kubernetes clusters for test runs.

Design constraints:

- Supports **`kind`** for CI environments (no external infrastructure required).
- Supports an **existing kubeconfig** for local development or staging against a real cluster.
- The framework does **not** own the cluster implementation — it only needs a valid kubeconfig to operate.

**Gap:** Whether `k3d` or other providers beyond `kind` and kubeconfig are first-class is mentioned in the diagram but not elaborated in the design text.

### Agent Manager

The **Agent Manager** deploys, restarts, and kills agents within the test cluster.

Deployment modes:

| Mode | Use case |
|------|----------|
| **Pods in cluster** | Production-like testing; agents run as Kubernetes workloads. |
| **Local processes** | Faster iteration during development. |

Controls exposed for scenario composition:

- **Start/stop individual agents mid-scenario** — For example, to test restart behavior, leader election, or recovery after crash.
- **Inject resource limits or network policies** — To simulate degraded runtime conditions.

**Gap:** API surface for these controls (REST, in-process calls, YAML hooks) is not defined. Container image sources and agent configuration are not specified.

### Scenario Engine

The **Scenario Engine** executes a single test scenario end to end:

1. **Apply initial state** — From Kubernetes manifests (YAML files) or programmatic resource creation.
2. **Fire the trigger** — Apply the mutation, restart, or fault defined in the scenario.
3. **Poll the cluster** — Until expected state is reached or timeout expires.
4. **Record outcome** — Pass or fail, with diagnostics collected on failure.

The engine is the component that interprets scenario YAML and drives assertions.

---

## Scenario definition

Scenarios are YAML files. The design defines the following structure:

| Field | Purpose |
|-------|---------|
| `name` | Identifier for the scenario. |
| `description` | Human-readable explanation of what is being tested. |
| `agents` | List of agent names in the Agent Set. |
| `setup.manifests` | Paths to Kubernetes manifest files for initial cluster state. |
| `trigger` | Optional event that provokes agent activity (for example, a resource patch). |
| `expect` | List of state assertions with resource selectors and conditions. |
| `timeout` | Maximum wait time for convergence (for example, `120s`). |

### Example scenario

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

### Trigger types (design)

The example shows a **patch** trigger. The design also states triggers may include:

- Resource mutation (as in the example)
- Agent restart
- Fault injection (see [Fault injection](#fault-injection))

**Gap:** YAML schema for non-patch triggers, programmatic setup (alternative to `setup.manifests`), and validation rules are not yet defined.

### Assertion conditions

Each `expect` entry selects a resource by `apiVersion`, `kind`, `name`, and `namespace`, then lists `conditions` as path/value pairs. Paths use JSONPath-style notation (for example, `.spec.replicas`).

**Gap:** Support for asserting resource absence, label selectors, multiple resources, or custom condition operators is not described in the design.

---

## Scope

### In scope

The framework is intended to test:

- **Agent-to-agent interaction** — Coordination, conflict resolution, and ordering when multiple agents act on shared resources.
- **Convergence under normal and degraded conditions** — Correct final cluster state when agents operate under typical or impaired runtime conditions.
- **Recovery after agent restart or crash** — Behavior when agents are killed or restarted mid-scenario.
- **Correct final state of cluster resources** — Field values, conditions, and readiness as declared in assertions.

### Out of scope

The framework explicitly does **not** target:

- **Agent internal logic** — Use unit tests for individual agent code paths.
- **Performance and load** — Treated as a separate concern.
- **Kubernetes correctness** — The cluster API and controllers are assumed to work correctly; tests focus on agent behavior.

---

## Failure diagnostics

When a scenario fails (timeout expires before expected state is reached), the framework collects diagnostic material so failures can be understood without manual reproduction.

| Diagnostic | Description |
|------------|-------------|
| **Agent logs** | Filtered to the scenario's namespace and relevant resources. |
| **Kubernetes events** | Events in the test namespace during the run. |
| **Resource diff** | Difference between expected and actual resource state at failure time. |
| **Mutation timeline** | Chronological record of resource changes from a watch stream captured during the test. |

Together, these artifacts aim to answer *why* the cluster did not converge — for example, whether an agent never acted, acted incorrectly, or was blocked by another agent or cluster condition.

**Gap:** Output location (files, stdout, structured report), log retention, and redaction policies are not specified.

---

## Fault injection

Optional **fault hooks** can be composed into scenarios to test resilience and edge-case behavior. The design defines the following faults:

| Fault | Mechanism | Purpose |
|-------|-----------|---------|
| **Kill agent** | Delete pod or kill process | Test recovery and leader re-election |
| **Network partition** | NetworkPolicy between agent and API server | Test behavior when the agent cannot reach the cluster |
| **Slow API server** | Inject latency via proxy | Test timeout and retry logic |
| **Stale cache** | Restart informer without full resync | Test correctness with partial state |
| **Resource conflict** | Concurrent update from test harness | Test conflict retry logic |

Faults may appear as scenario triggers or as mid-scenario actions coordinated by the Agent Manager or Scenario Engine.

**Gap:** YAML syntax for declaring faults, ordering relative to triggers, and composability of multiple simultaneous faults are not defined.

---

## Implementation status

The framework is in the **design phase**. The following implementation milestones are planned but not yet built:

1. Cluster provider with `kind` support
2. Agent manager that deploys agents from container images
3. Scenario engine: YAML parsing, state application, polling assertions
4. Failure diagnostics collection
5. CLI: `kube-agents-test run scenarios/`
6. CI integration (GitHub Actions workflow)

### Technology choices

These choices are fixed in the design and inform future implementation:

| Choice | Rationale |
|--------|-----------|
| **Go** | Same language as the agents; shared `client-go` usage; no FFI boundary. |
| **client-go** | Direct Kubernetes API interaction, watches, dynamic client for arbitrary resources. |
| **kind** | Ephemeral clusters in CI without infrastructure dependencies. |
| **No test framework dependency** | Scenarios are data; the runner is a standalone binary, not a `go test` suite. Enables running scenarios against any cluster. |

---

## Related reading

- [Documentation index](README.md) — Overview and documentation structure.
