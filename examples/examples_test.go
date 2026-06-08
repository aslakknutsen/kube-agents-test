package examples_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExamplesBuild(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(root) == "examples" {
		root = filepath.Dir(root)
	}

	examples := []string{
		"./examples/load-scenario",
		"./examples/run-with-stubs",
		"./examples/run-with-fakes",
		"./examples/run-attached",
		"./examples/run-suite",
	}
	for _, pkg := range examples {
		pkg := pkg
		t.Run(pkg, func(t *testing.T) {
			cmd := exec.Command("go", "build", "-o", os.DevNull, pkg)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("build %s: %v\n%s", pkg, err, out)
			}
		})
	}
}

func TestExampleScenariosLoad(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(root) == "examples" {
		root = filepath.Dir(root)
	}

	scenarios := []string{
		"examples/scenarios/scaling-quota/scenario.yaml",
		"examples/scenarios/reconcile-only/scenario.yaml",
		"examples/scenarios/agent-restart/scenario.yaml",
		"examples/scenarios/fault-kill-agent/scenario.yaml",
	}
	for _, rel := range scenarios {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			cmd := exec.Command("go", "run", "./examples/load-scenario", rel)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("load %s: %v\n%s", rel, err, out)
			}
		})
	}
}
