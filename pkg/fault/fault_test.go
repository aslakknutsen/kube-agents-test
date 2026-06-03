package fault_test

import (
	"testing"
	"time"

	"github.com/kube-agents/kube-agents-test/pkg/fault"
	"gopkg.in/yaml.v3"
)

func TestSpecUnmarshalDelay(t *testing.T) {
	var spec fault.Spec
	if err := yaml.Unmarshal([]byte(`kind: slowAPIServer
delay: 500ms
`), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Delay != 500*time.Millisecond {
		t.Fatalf("delay = %v", spec.Delay)
	}
}

func TestSpecUnmarshalResourceConflict(t *testing.T) {
	var spec fault.Spec
	if err := yaml.Unmarshal([]byte(`kind: resourceConflict
patch:
  apiVersion: apps/v1
  kind: Deployment
  name: target
  namespace: test
  spec:
    replicas: 99
`), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Kind != fault.KindResourceConflict {
		t.Fatalf("kind = %q", spec.Kind)
	}
	if spec.Patch == nil || spec.Patch.Spec["replicas"] != 99 {
		t.Fatalf("patch = %+v", spec.Patch)
	}
}
