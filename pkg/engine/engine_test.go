package engine

import (
	"testing"
)

func TestLookupPath(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": int64(3),
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "nginx",
					},
				},
			},
		},
		"status": map[string]interface{}{
			"readyReplicas": float64(3),
		},
	}

	tests := []struct {
		name     string
		path     string
		wantVal  interface{}
		wantFind bool
	}{
		{"nested int", ".spec.replicas", int64(3), true},
		{"deep nested string", ".spec.template.metadata.labels.app", "nginx", true},
		{"float value", ".status.readyReplicas", float64(3), true},
		{"missing key", ".spec.missing", nil, false},
		{"missing deep", ".spec.template.missing.field", nil, false},
		{"no leading dot", "spec.replicas", int64(3), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, found, err := lookupPath(obj, tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found != tt.wantFind {
				t.Errorf("found = %v, want %v", found, tt.wantFind)
			}
			if found && val != tt.wantVal {
				t.Errorf("val = %v (%T), want %v (%T)", val, val, tt.wantVal, tt.wantVal)
			}
		})
	}

	// Test that a top-level map path returns non-nil.
	t.Run("top-level map", func(t *testing.T) {
		val, found, err := lookupPath(obj, ".spec")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Error("expected .spec to be found")
		}
		m, ok := val.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map, got %T", val)
		}
		if _, has := m["replicas"]; !has {
			t.Error("expected spec map to contain 'replicas'")
		}
	})
}

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
		want     bool
	}{
		{"int vs int", 5, 5, true},
		{"float64 vs int", float64(5), 5, true},
		{"int vs float64", 5, float64(5), true},
		{"int64 vs int", int64(5), 5, true},
		{"string vs string", "hello", "hello", true},
		{"string mismatch", "hello", "world", false},
		{"int mismatch", 3, 5, false},
		{"float64 mismatch", float64(3), 5, false},
		{"bool true", true, true, true},
		{"bool false mismatch", true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.actual, tt.expected, got, tt.want)
			}
		})
	}
}
