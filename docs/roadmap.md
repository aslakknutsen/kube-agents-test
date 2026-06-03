# Roadmap and technology choices

This page records the **planned implementation order** and **technology choices** from the project design. It is a roadmap, not a guarantee of current code availability — refer to the repository for what is implemented today.

## Implementation plan

Work is intended to proceed in this order:

| Step | Deliverable |
|------|-------------|
| 1 | **Cluster provider** — ephemeral clusters (`kind`) for CI and attached clusters (kubeconfig), including full platforms such as OpenShift |
| 2 | **Agent manager** — deploy agents from container images (and local process mode where applicable) |
| 3 | **Scenario engine** — YAML parsing, initial state application, polling assertions |
| 4 | **Failure diagnostics** — logs, events, diffs, mutation timeline on failure |
| 5 | **CLI** — `kube-agents-test run scenarios/` to execute scenario files or directories |
| 6 | **CI integration** — GitHub Actions workflow running scenarios against kind (or equivalent) |

Earlier steps unblock later ones: without a cluster and agents, the engine cannot run meaningful scenarios; without diagnostics, failed CI runs are harder to triage.

## Technology choices

| Choice | Rationale |
|--------|-----------|
| **Go** | Same language as the agents; shared patterns and types; no foreign-function boundary between test harness and agent code |
| **client-go** | Direct Kubernetes API access, watches, and dynamic client for arbitrary resources in assertions |
| **kind** | Ephemeral clusters in CI without depending on external cloud infrastructure |
| **No test framework dependency** | Scenarios are **data** (YAML), not Go test cases. The runner is a **standalone binary**, not embedded in `go test`. Decouples scenario execution from Go's test semantics and allows running the same files against **any** cluster with a suitable kubeconfig |

## Architectural outcomes

When the plan is complete, users should be able to:

1. Point the tool at a cluster (provisioned by kind in CI or existing kubeconfig locally).
2. Declare scenarios as YAML under `scenarios/` (or similar).
3. Run the CLI in CI and locally with the same files.
4. Rely on eventual-consistency assertions and rich failure output for multi-agent debugging.

Conceptual building blocks are documented in [Core concepts](core-concepts.md) and [Architecture](architecture.md).

## Documentation map

| Topic | Document |
|-------|----------|
| Test Scenario, Agent Set, State Assertion | [Core concepts](core-concepts.md) |
| Test Runner, providers, engine | [Architecture](architecture.md) |
| YAML format and example | [Scenarios](scenarios.md) |
| In/out of scope | [Scope](scope.md) |
| Failure artifacts | [Failure diagnostics](failure-diagnostics.md) |
| Fault catalog | [Fault injection](fault-injection.md) |
