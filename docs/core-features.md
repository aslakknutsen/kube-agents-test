# Core Framework Features

kube-agents-test is a high-level testing framework for the kube-agents platform — a system where multiple autonomous agents operate on a shared Kubernetes cluster.

## Problem statement

Unit testing individual agents is straightforward. The hard part is testing the *system*: multiple agents reacting to the same cluster state, potentially conflicting, racing, or depending on each other's outputs. High-level tests verify that agents, taken together, drive the cluster toward the correct state.

---

## Core concepts

### Test Scenario

A **Test Scenario** is a declarative description of an end-to-end multi-agent test. It defines four elements:

1. **Initial cluster state** — Resources to pre-create before agents act (manifests or programmatic creation).
2. **Optional trigger** — An event that starts or perturbs the scenario: a resource mutation, agent restart, or fault injection.
3. **Expected final state** — Resource conditions, field values, or resource absence that indicate success.
4. **Timeout** — Maximum time to wait for convergence before the scenario fails.

Scenarios are data (YAML files), not Go test functions. This keeps tests portable across clusters and decoupled from `go test` semantics.

**Purpose:** Encode a reproducible multi-agent interaction as declarative input the runner can execute without custom code per test.

**Behavior:** The Scenario Engine loads a scenario, applies setup, fires the trigger, and evaluates assertions until pass, fail, or timeout.

**Relationships:** Consumed by the Scenario Engine; references an Agent Set; defines State Assertions under `expect`.

---

### Agent Set

An **Agent Set** specifies which agents participate in a scenario.

- Run a **subset** of agents to isolate interactions (e.g., only scaling-agent and quota-agent).
- Run the **full set** for true integration tests across all platform agents.

**Purpose:** Control blast radius and test focus — isolate two-agent conflicts or exercise the whole system.

**Behavior:** The Agent Manager deploys and manages only the agents listed in the scenario's `agents` field.

**Relationships:** Declared in scenario YAML; drives Agent Manager deployment and log filtering during failure diagnostics.

---

### State Assertion

A **State Assertion** is a polling- or watch-based check that waits for the cluster to converge to the expected state within the scenario timeout.

Assertions are **not** point-in-time snapshots. Agents are eventually consistent, so the framework must wait for convergence rather than sampling once immediately after a trigger.

**Purpose:** Match how distributed agents actually behave — success means the cluster reached the expected steady state, not that it happened instantly.

**Behavior:** The Scenario Engine polls or watches resources until all conditions in `expect` match or the timeout expires.

**Relationships:** Defined under `expect` in scenario YAML; evaluated by the Scenario Engine after setup and trigger.

---

## Architecture

The framework is organized around a Test Runner that orchestrates three subsystems:

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

**Purpose:** Top-level orchestrator for test lifecycle and result collection.

**Behavior:** Coordinates Cluster Provider, Agent Manager, and Scenario Engine; aggregates pass/fail and diagnostics across scenarios.

**Relationships:** Invoked by the CLI (`kube-agents-test run scenarios/`); owns cluster and agent lifecycle for a test run.

---

### Cluster Provider

**Purpose:** Supply a Kubernetes API endpoint (kubeconfig) for the test run.

**Behavior:**

- Creates and tears down **ephemeral clusters** for isolated CI runs.
- Supports **`kind`** for CI environments without external infrastructure.
- Accepts an **existing kubeconfig** for dev or staging against a pre-provisioned cluster.

The framework does not own the cluster implementation — it only needs a valid kubeconfig to drive the Scenario Engine and Agent Manager.

**Relationships:** Used by Test Runner before scenarios run; teardown happens after the test suite completes.

---

### Agent Manager

**Purpose:** Deploy, restart, and kill agents within the test cluster.

**Behavior:**

- Deploy agents as **pods** (production-like) or run them as **local processes** (faster iteration).
- **Start/stop individual agents** mid-scenario to test restart behavior, leader election, and recovery.
- **Inject degraded conditions** via resource limits or network policies.

**Relationships:** Receives the Agent Set from the scenario; operates on the cluster from Cluster Provider; supports fault injection hooks (e.g., kill agent).

---

### Scenario Engine

**Purpose:** Execute a single test scenario from setup through assertion.

**Behavior:**

1. **Apply initial state** — Kubernetes manifests or programmatic resource creation from `setup`.
2. **Fire the trigger** — Patch, agent action, or fault injection from `trigger`.
3. **Poll the cluster** — Watch or poll until expected state is reached or timeout expires.
4. **Record pass/fail** — On failure, collect diagnostics (agent logs, resource diffs, events).

**Relationships:** Orchestrates Cluster Provider (cluster access), Agent Manager (agent lifecycle), and State Assertions (`expect`). Primary executor of scenario YAML.

---

## Scenario definition format

Scenarios are YAML files. The following example is taken from the framework design:

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

### Field reference

| Section | Role |
|---------|------|
| `name` | Unique scenario identifier |
| `description` | Human-readable intent |
| `agents` | Agent Set for this scenario |
| `setup.manifests` | Initial cluster state (manifest paths) |
| `trigger` | Optional event that perturbs the cluster (e.g., `patch`) |
| `expect` | State Assertions: resource identity, conditions, timeout |
| `expect[].conditions` | JSONPath-style paths and expected values |
| `timeout` | Convergence deadline for assertions |

Scenarios are **data, not code**. The runner is a standalone binary, not a `go test` suite, so scenarios can run against any cluster given a kubeconfig.

---

## Failure diagnostics

When a scenario fails, the framework collects enough context to debug *why* the cluster did not converge without manual reproduction:

| Diagnostic | Description |
|------------|-------------|
| **Agent logs** | Filtered to the scenario's namespace and relevant resources |
| **Kubernetes events** | Events in the test namespace during the run |
| **Resource diff** | Expected vs. actual state at failure time |
| **Mutation timeline** | Resource changes recorded from a watch stream during the test |

**Purpose:** Shorten the debug loop for flaky or ordering-dependent multi-agent failures.

**Behavior:** Collected automatically on assertion timeout or mismatch; attached to the scenario result by the Scenario Engine / Test Runner.

---

## Fault injection

Optional fault hooks can be composed into scenarios to exercise resilience and edge cases:

| Fault | Mechanism | Purpose |
|-------|-----------|---------|
| Kill agent | Delete pod / kill process | Test recovery and leader re-election |
| Network partition | NetworkPolicy between agent and API server | Test agent behavior when it can't reach the cluster |
| Slow API server | Inject latency via proxy | Test timeout and retry logic |
| Stale cache | Restart informer without full resync | Test agent correctness with partial state |
| Resource conflict | Concurrent update from test harness | Test conflict retry logic |

**Purpose:** Simulate degraded or adversarial conditions beyond happy-path convergence.

**Behavior:** Invoked as part of `trigger` or mid-scenario via Agent Manager / harness controls; scenarios remain declarative where possible.

**Relationships:** Agent Manager executes agent-level faults; cluster-level faults use Kubernetes primitives (NetworkPolicy, etc.) or harness-side proxies.

---

## Scope boundaries

### In scope

- **Agent-to-agent interaction** — Coordination, conflict resolution, ordering between agents.
- **Convergence under normal and degraded conditions** — Steady-state correctness when the system is healthy or partially impaired.
- **Recovery after agent restart or crash** — Scenarios that kill or restart agents mid-run.
- **Correct final state of cluster resources** — Assertions on resource fields, conditions, and presence/absence.

### Out of scope

- **Agent internal logic** — Use unit tests for single-agent behavior.
- **Performance and load** — Separate concern; not covered by this framework.
- **Kubernetes correctness** — The framework assumes Kubernetes behaves correctly.

---

## Implementation plan

Planned delivery order:

1. **Cluster provider** with `kind` support
2. **Agent manager** that deploys agents from container images
3. **Scenario engine** — YAML parsing, state application, polling assertions
4. **Failure diagnostics** collection
5. **CLI** to run scenarios (`kube-agents-test run scenarios/`)
6. **CI integration** — GitHub Actions workflow

---

## Technology choices

| Choice | Rationale |
|--------|-----------|
| **Go** | Same language as the agents; shared `client-go` usage; no FFI boundary |
| **client-go** | Direct Kubernetes API interaction, watches, dynamic client for arbitrary resources |
| **kind** | Ephemeral clusters in CI without external infrastructure dependencies |
| **No test framework dependency** | Scenarios are data; runner is a standalone binary decoupled from `go test` |

---

## Component interaction summary

For a single scenario run:

1. **Test Runner** obtains a cluster from **Cluster Provider** (kubeconfig).
2. **Agent Manager** deploys the scenario's **Agent Set**.
3. **Scenario Engine** applies **setup**, fires **trigger**, runs **State Assertions** under **expect**.
4. On failure, diagnostics (logs, events, diffs, timeline) are collected.
5. Optional **fault injection** hooks perturb agents or the cluster during steps 2–3.
6. **Test Runner** records the result and proceeds to teardown.

This flow ties the core concepts to the architectural components without requiring scenario authors to manage low-level cluster or agent lifecycle details.
