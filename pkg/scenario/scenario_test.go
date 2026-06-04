package scenario_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestLoadFile_exampleFromDocs(t *testing.T) {
	path := filepath.Join("testdata", "example-scenario.yaml")
	sc, err := scenario.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if sc.Name != "scaling-agent-respects-quota-agent" {
		t.Errorf("name = %q", sc.Name)
	}
	if len(sc.Agents) != 2 {
		t.Fatalf("agents len = %d", len(sc.Agents))
	}
	if len(sc.Setup.Manifests) != 2 {
		t.Fatalf("setup.manifests len = %d", len(sc.Setup.Manifests))
	}
	if sc.Trigger == nil || sc.Trigger.Patch == nil {
		t.Fatal("expected trigger.patch")
	}
	if sc.Trigger.Patch.Kind != "Deployment" {
		t.Errorf("patch kind = %q", sc.Trigger.Patch.Kind)
	}
	if sc.Trigger.Patch.Body["spec"] == nil {
		t.Error("expected patch spec body")
	}
	if sc.Expect.Timeout != 120*time.Second {
		t.Errorf("timeout = %s", sc.Expect.Timeout)
	}
	if len(sc.Expect.Resources) != 1 {
		t.Fatalf("expect resources len = %d", len(sc.Expect.Resources))
	}
	re := sc.Expect.Resources[0]
	if len(re.Conditions) != 2 {
		t.Fatalf("conditions len = %d", len(re.Conditions))
	}
	if re.Conditions[0].Path != ".spec.replicas" {
		t.Errorf("path = %q", re.Conditions[0].Path)
	}
}

func TestValidate_requiresFields(t *testing.T) {
	_, err := scenario.LoadFile(filepath.Join("testfixtures", "invalid-missing-timeout.yaml"))
	if err == nil {
		t.Fatal("expected validation error for missing timeout")
	}
}
