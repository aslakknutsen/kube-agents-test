package scenario

import (
	"gopkg.in/yaml.v3"
)

// UnmarshalYAML decodes apiVersion/kind/name/namespace and remaining keys into Body.
func (p *ResourcePatch) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]any
	if err := value.Decode(&raw); err != nil {
		return err
	}
	p.APIVersion, _ = raw["apiVersion"].(string)
	p.Kind, _ = raw["kind"].(string)
	p.Name, _ = raw["name"].(string)
	if ns, ok := raw["namespace"].(string); ok {
		p.Namespace = ns
	}
	p.Body = make(map[string]any)
	for _, k := range []string{"apiVersion", "kind", "name", "namespace"} {
		delete(raw, k)
	}
	for k, v := range raw {
		p.Body[k] = v
	}
	return nil
}

// MarshalYAML writes ObjectRef fields and patch body together.
func (p ResourcePatch) MarshalYAML() (any, error) {
	out := make(map[string]any)
	if p.APIVersion != "" {
		out["apiVersion"] = p.APIVersion
	}
	if p.Kind != "" {
		out["kind"] = p.Kind
	}
	if p.Name != "" {
		out["name"] = p.Name
	}
	if p.Namespace != "" {
		out["namespace"] = p.Namespace
	}
	for k, v := range p.Body {
		out[k] = v
	}
	return out, nil
}

