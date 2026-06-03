# API usage examples

Runnable programs that show how to wire the public `pkg/*` APIs. They use in-memory fakes (not kind or client-go) so they build and run anywhere.

| Example | What it demonstrates |
|---------|----------------------|
| [scenario_load](scenario_load/) | `scenario.Load`, `Validate`, `ResolveManifestPaths` |
| [runner](runner/) | `runner.New`, `Config`, `Dependencies`, `Run` / `RunPath`, result summary |
| [engine](engine/) | `engine.ExecuteRequest` and inspecting `engine.Result` |

```bash
go run ./examples/scenario_load
go run ./examples/runner
go run ./examples/engine
```

When real backends land, swap the fakes for `internal/cluster/kind`, a concrete `agent.Manager`, and a real `engine.Engine` — the `pkg` types stay the same.
