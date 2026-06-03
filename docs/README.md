# kube-agents-test — User documentation

This folder contains user-facing documentation extracted from the project design. It describes what the framework does, how scenarios are defined, and how the pieces fit together. Implementation details may evolve; behavior described here matches the current design in the repository README.

## Contents

| Document | Description |
|----------|-------------|
| [Core concepts](core-concepts.md) | Test Scenario, Agent Set, and State Assertion |
| [Architecture](architecture.md) | Test Runner, Cluster Provider, Agent Manager, and Scenario Engine |
| [Scenarios](scenarios.md) | Declarative YAML scenario format and lifecycle |
| [Scope](scope.md) | What the framework tests and what it does not |
| [Failure diagnostics](failure-diagnostics.md) | Artifacts collected when a scenario fails |
| [Fault injection](fault-injection.md) | Optional faults composable into scenarios |
| [Roadmap](roadmap.md) | Planned implementation order and technology choices |

## Audience

These docs are for anyone writing or running high-level tests against the kube-agents platform: verifying that multiple autonomous agents, together, drive a shared Kubernetes cluster toward the correct state.

## Relationship to the README

The repository [README](../README.md) is the canonical high-level design summary. This documentation elaborates the same material for inclusion in a fuller user guide later.
