# Failure diagnostics

When a scenario **does not converge** within its timeout — expected state never matched, or matched too late — the run fails. The framework then collects structured diagnostics so you can see **why** the cluster did not reach the expected state without manually reproducing the run.

Diagnostics are gathered by the [Scenario Engine](architecture.md#scenario-engine) and [Test Runner](architecture.md#test-runner) as part of the failure path, not as a separate manual step.

## Collected artifacts

| Artifact | Description |
|----------|-------------|
| **Agent logs** | Log output from agents involved in the scenario, **filtered** to the scenario's namespace and relevant resources so noise from unrelated workloads is reduced |
| **Kubernetes events** | Events in the **test namespace** — scheduling failures, quota denials, warnings from controllers, and other signals that explain resource state |
| **Expected vs actual diff** | Comparison between what the scenario's `expect` section required and what was observed on the cluster at failure time (field-level mismatch) |
| **Mutation timeline** | Chronological record of resource changes from a **watch stream** captured during the test — shows ordering of patches, status updates, and deletions while agents reconciled |

Together these answer:

- Did agents log errors or retries?
- Did the API server or controllers emit events explaining stuck state?
- Which fields diverged from the assertion?
- What sequence of changes occurred before timeout?

## When diagnostics run

Diagnostics are collected **on failure** after polling assertions exhaust the scenario timeout or detect an unrecoverable mismatch. Successful runs do not require this bundle for pass/fail determination; the design optimizes for **debuggability of failures** in multi-agent, eventually consistent environments.

## Using diagnostics effectively

1. **Start with the diff** — Identifies which `expect` conditions failed (e.g. replica count, ready replicas).
2. **Check events** — Often explains *why* resources stayed pending (quota, image pull, policy).
3. **Read filtered agent logs** — Shows reconciliation errors, conflicts, or leader transitions for the agents in the [agent set](core-concepts.md#agent-set).
4. **Review the timeline** — Resolves races: which agent wrote last, whether a trigger was applied before an agent restarted, etc.

This matches the design goal: enough context to debug non-convergence **in place**, rather than re-running with ad hoc `kubectl` commands alone.

## Relationship to scenarios

Scenario YAML defines **what** should have happened (`expect`) and **how long** to wait (`timeout`). Diagnostics explain **what actually happened** when that contract was not met. See [Scenarios](scenarios.md) for defining expectations and [Scope](scope.md) for what class of bugs scenarios are meant to catch.
