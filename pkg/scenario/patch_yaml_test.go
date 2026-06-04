package scenario_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

func TestResourcePatch_unmarshalSeparatesRefAndBody(t *testing.T) {
	const raw = `
apiVersion: apps/v1
kind: Deployment
name: target
namespace: test
spec:
  replicas: 3
`
	var p scenario.ResourcePatch
	if err := yaml.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	if p.APIVersion != "apps/v1" || p.Kind != "Deployment" || p.Name != "target" || p.Namespace != "test" {
		t.Fatalf("ref = %+v", p.ObjectRef)
	}
	spec, ok := p.Body["spec"].(map[string]any)
	if !ok {
		t.Fatalf("body spec = %#v", p.Body["spec"])
	}
	if spec["replicas"] != 3 {
		t.Errorf("replicas = %v", spec["replicas"])
	}
	if _, ok := p.Body["apiVersion"]; ok {
		t.Error("apiVersion should not remain in Body")
	}
}

func TestResourcePatch_marshalRoundTrip(t *testing.T) {
	in := scenario.ResourcePatch{
		ObjectRef: scenario.ObjectRef{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Name:       "c",
		},
		Body: map[string]any{"data": map[string]any{"k": "v"}},
	}
	data, err := yaml.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out scenario.ResourcePatch
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != in.Name || out.Body["data"] == nil {
		t.Fatalf("out = %+v", out)
	}
}
