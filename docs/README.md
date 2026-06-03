# kube-agents-test — User documentation

Documentation for the **kube-agents-test** high-level testing framework: a system for verifying that multiple autonomous agents, operating on a shared Kubernetes cluster, drive the cluster toward the correct state together.

This folder is the home for user-facing docs. Content is derived from the project [design README](../README.md) until implementation exists.

## Contents

| Document | Description |
|----------|-------------|
| [Core framework features](core-features.md) | Concepts, architecture, scenario model, scope, diagnostics, fault injection, and technology choices |

## Audience

- Engineers writing or running multi-agent integration scenarios
- Contributors implementing the framework (cluster provider, agent manager, scenario engine)

## Source of truth

Until code lands in the repository, the authoritative design description is [README.md](../README.md) at the repository root. User docs here elaborate that design without adding APIs or behavior not stated there.
