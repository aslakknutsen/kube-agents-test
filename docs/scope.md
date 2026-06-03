# Scope

kube-agents-test targets **system-level** behavior: multiple agents sharing one cluster and producing a correct **final** cluster state. This page clarifies what belongs in framework scenarios versus other test layers.

## In scope

Use scenarios when you need to verify:

| Area | What you are testing |
|------|----------------------|
| **Agent-to-agent interaction** | Coordination, conflict resolution, and ordering when multiple agents touch the same resources |
| **Convergence** | The cluster reaches the expected state under normal and **degraded** conditions (limits, policies, partial failures) |
| **Recovery** | Correct behavior after **agent restart or crash** (leader election, resync, catch-up reconciliation) |
| **Final resource state** | Conditions and fields on Kubernetes objects (or their absence) after agents finish reconciling |

These map directly to the framework's [core concepts](core-concepts.md): declarative scenarios, selectable agent sets, and eventual-consistency assertions with timeouts.

## Out of scope

Do **not** rely on this framework for:

| Area | Where to test instead |
|------|------------------------|
| **Agent internal logic** | Unit tests in each agent's codebase |
| **Performance and load** | Dedicated performance or soak testing; not the scenario runner's goal |
| **Kubernetes platform correctness** | Assumed correct; scenarios test *your agents'* use of the API, not upstream kube bugs |

Keeping this boundary clear avoids slow, flaky scenarios that duplicate unit tests or require benchmarking infrastructure.

## Problem framing

**Unit testing** individual agents is straightforward. The hard problem is the **system**: agents reacting to the same cluster state, potentially **conflicting**, **racing**, or depending on **each other's outputs**. High-level tests assert that the agents, **taken together**, drive the cluster toward the correct state — which is what YAML scenarios and the [Scenario Engine](architecture.md#scenario-engine) are for.

## Related capabilities

Capabilities that support in-scope testing but are not separate "products":

- **[Fault injection](fault-injection.md)** — Injects kills, partitions, latency, stale caches, and conflicts to test convergence and recovery under stress.
- **[Failure diagnostics](failure-diagnostics.md)** — When convergence fails, collects logs, events, diffs, and timelines so failures are debuggable without manual reproduction.

For what to put in a scenario file, see [Scenarios](scenarios.md).
