# Go API surface

Public packages under `pkg/` define the test framework boundaries. Implementations follow the [roadmap](roadmap.md).

| Package | Role |
|---------|------|
| [`pkg/scenario`](../pkg/scenario/doc.go) | YAML scenario model, load, validate |
| [`pkg/fault`](../pkg/fault/doc.go) | Trigger and fault-injection types |
| [`pkg/cluster`](../pkg/cluster/doc.go) | Cluster provider interface and handle |
| [`pkg/agent`](../pkg/agent/doc.go) | Agent registry and manager interfaces |
| [`pkg/engine`](../pkg/engine/doc.go) | Scenario engine interface and results |
| [`pkg/diag`](../pkg/diag/doc.go) | Failure diagnostics collector |
| [`pkg/runner`](../pkg/runner/doc.go) | Test suite orchestration |

Canonical scenario `expect` shape:

```yaml
expect:
  timeout: 120s
  assertions:
    - resource: { ... }
      conditions: [ ... ]
```

The scenario loader also accepts the legacy sequence form from [scenarios.md](scenarios.md) during transition.
