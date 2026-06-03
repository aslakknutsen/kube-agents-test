// Scenario load example: read a YAML scenario, validate it, resolve manifest paths.
//
// Run: go run ./examples/scenario_load
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

const sampleScenario = `name: scaling-agent-respects-quota-agent
description: Agents converge under quota when scaling is requested.

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
      replicas: 10

expect:
  - resource:
      apiVersion: apps/v1
      kind: Deployment
      name: target
      namespace: test
    conditions:
      - path: .spec.replicas
        value: 5
  timeout: 120s
`

func main() {
	dir, err := os.MkdirTemp("", "kube-agents-scenario-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "scenario.yaml")
	if err := os.WriteFile(path, []byte(sampleScenario), 0o644); err != nil {
		log.Fatal(err)
	}

	sc, err := scenario.Load(path)
	if err != nil {
		log.Fatalf("Load: %v", err)
	}

	if err := sc.Validate(); err != nil {
		log.Fatalf("Validate: %v", err)
	}

	base := scenario.BaseDir(path)
	manifests, err := sc.ResolveManifestPaths(base)
	if err != nil {
		log.Fatalf("ResolveManifestPaths: %v", err)
	}

	fmt.Printf("scenario: %s\n", sc.Name)
	fmt.Printf("agents: %v\n", sc.Agents)
	fmt.Printf("expect timeout: %s\n", sc.Expect.Timeout)
	fmt.Printf("assertions: %d\n", len(sc.Expect.Assertions))
	if sc.Trigger != nil && sc.Trigger.Patch != nil {
		fmt.Printf("trigger patch: %s/%s in %s\n",
			sc.Trigger.Patch.Kind, sc.Trigger.Patch.Name, sc.Trigger.Patch.Namespace)
	}
	fmt.Printf("manifest paths (relative to %s):\n", base)
	for _, m := range manifests {
		fmt.Printf("  - %s\n", m)
	}
}
