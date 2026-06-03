# kube-agents-test User Documentation

User documentation for **kube-agents-test**, a high-level testing framework for the kube-agents platform — a system where multiple autonomous agents operate on a shared Kubernetes cluster.

## What this documentation covers

This documentation is extracted from the project design and will grow as the framework is implemented. It describes what the framework is intended to do, how its pieces fit together, and how you will define and run test scenarios.

## Contents

| Document | Description |
|----------|-------------|
| [Core Features](core-features.md) | Core concepts, architecture, scenario format, scope, diagnostics, and fault injection |

## Problem the framework solves

Unit testing individual agents is straightforward. The hard part is testing the *system*: multiple agents reacting to the same cluster state, potentially conflicting, racing, or depending on each other's outputs. High-level tests need to verify that the agents, taken together, drive the cluster toward the correct state.

## Planned tooling

The implementation plan (not yet built) includes:

- A CLI to run scenarios: `kube-agents-test run scenarios/`
- CI integration via a GitHub Actions workflow

See [Core Features — Implementation status](core-features.md#implementation-status) for what is defined in design versus what remains to be built.
