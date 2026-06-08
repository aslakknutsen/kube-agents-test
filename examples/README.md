# API usage examples

Runnable examples for the kube-agents-test framework API. Each subdirectory is a standalone `main` program; scenario YAML lives under `scenarios/`.

## Prerequisites

- Go 1.22+
- From the repository root: `go build ./examples/...`

Subsystems (cluster provider, agent manager, scenario engine) are stubs in the current API scaffold — examples use either stub backends (expect `ErrNotImplemented`) or inline fakes to show orchestration without a real cluster.

## Go programs

| Example | What it demonstrates |
|---------|---------------------|
| [load-scenario](load-scenario/main.go) | Load a YAML file, inspect parsed fields, run `ValidateWith` |
| [run-with-stubs](run-with-stubs/main.go) | Minimal `runner.NewDefault` with stub dependencies |
| [run-with-fakes](run-with-fakes/main.go) | Inject fake provider/manager/engine to exercise full orchestration |
| [run-attached](run-attached/main.go) | Attached-cluster `RunOptions` (kubeconfig, sandbox, leave running) |
| [run-suite](run-suite/main.go) | Multi-scenario directory run with `FailFast` and diagnostics |

## Scenario YAML

Example scenario files under `scenarios/` map to documented YAML shapes:

| Scenario | Trigger type | Notes |
|----------|--------------|-------|
| [scaling-quota](scenarios/scaling-quota/scenario.yaml) | `patch` | Scaling vs quota interaction (from docs) |
| [reconcile-only](scenarios/reconcile-only/scenario.yaml) | none | Setup + expect only |
| [agent-restart](scenarios/agent-restart/scenario.yaml) | `agentRestart` | Mid-run agent restart hook |
| [fault-kill-agent](scenarios/fault-kill-agent/scenario.yaml) | `fault` | Requires `AllowFaults: true` at run time |

Manifest paths are relative to the scenario file directory; each scenario includes placeholder fixtures.

## Quick start

```bash
# Parse and validate a scenario without running tests
go run ./examples/load-scenario

# Run orchestration with stub backends (fails at cluster Ensure)
go run ./examples/run-with-stubs

# Run with fakes — completes successfully, no real cluster
go run ./examples/run-with-fakes

# Run all example scenarios (explicit paths — fixture YAML lives in subdirectories)
go run ./examples/run-suite
```

When real implementations land (roadmap steps 1–3), swap stub/fake dependencies for production backends in `runner.Dependencies` — `RunOptions` and scenario YAML stay the same.
