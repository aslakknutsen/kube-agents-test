// Load and validate a scenario YAML file without provisioning a cluster.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func main() {
	scenarioPath := filepath.Join("examples", "scenarios", "scaling-quota", "scenario.yaml")
	if len(os.Args) > 1 {
		scenarioPath = os.Args[1]
	}

	s, err := scenario.Load(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("name:        %s\n", s.Name)
	fmt.Printf("description: %s\n", s.Description)
	fmt.Printf("agents:      %v\n", s.Agents)
	fmt.Printf("manifests:   %v\n", s.Setup.Manifests)
	fmt.Printf("timeout:     %s\n", s.Expect.Timeout)
	fmt.Printf("assertions:  %d\n", len(s.Expect.Assertions))

	if s.Trigger != nil && s.Trigger.Patch != nil {
		p := s.Trigger.Patch
		fmt.Printf("trigger:     patch %s/%s %s/%s\n", p.APIVersion, p.Kind, p.Namespace, p.Name)
	}

	valCtx := scenario.ValidationContext{
		SandboxNamespace: "test-agents",
		AllowFaults:      s.Trigger != nil && s.Trigger.Fault != nil,
	}
	if err := s.ValidateWith(valCtx); err != nil {
		fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("validation:  ok (sandbox=test-agents)")
}
