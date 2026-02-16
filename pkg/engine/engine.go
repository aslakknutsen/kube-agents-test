// Package engine implements the scenario execution logic.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kube-agents/kube-agents-test/pkg/agent"
	"github.com/kube-agents/kube-agents-test/pkg/diagnostics"
	"github.com/kube-agents/kube-agents-test/pkg/scenario"
)

// Result holds the outcome of running a single scenario.
type Result struct {
	Scenario    string
	Passed      bool
	Duration    time.Duration
	Error       error
	Diagnostics *diagnostics.Report
}

// Engine executes test scenarios against a Kubernetes cluster.
type Engine struct {
	KubeconfigPath string
	AgentManager   agent.Manager
	Collector      diagnostics.Collector

	dynamicClient dynamic.Interface
	mapper        meta.RESTMapper
}

// ClientOverride lets tests inject a fake dynamic client and mapper.
func (e *Engine) ClientOverride(dc dynamic.Interface, mapper meta.RESTMapper) {
	e.dynamicClient = dc
	e.mapper = mapper
}

func (e *Engine) ensureClients() error {
	if e.dynamicClient != nil && e.mapper != nil {
		return nil
	}

	config, err := clientcmd.BuildConfigFromFlags("", e.KubeconfigPath)
	if err != nil {
		return fmt.Errorf("building kubeconfig: %w", err)
	}

	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("creating dynamic client: %w", err)
	}
	e.dynamicClient = dc

	disc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return fmt.Errorf("creating discovery client: %w", err)
	}
	groupResources, err := restmapper.GetAPIGroupResources(disc)
	if err != nil {
		return fmt.Errorf("discovering API resources: %w", err)
	}
	e.mapper = restmapper.NewDiscoveryRESTMapper(groupResources)

	return nil
}

// Run executes a single scenario and returns the result.
func (e *Engine) Run(ctx context.Context, s *scenario.Scenario) *Result {
	start := time.Now()
	result := &Result{Scenario: s.Name}

	timeout := s.Timeout.Duration
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := e.ensureClients(); err != nil {
		result.Error = fmt.Errorf("initializing clients: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// 1. Deploy agents.
	for _, name := range s.Agents {
		if err := e.AgentManager.Deploy(ctx, name); err != nil {
			result.Error = fmt.Errorf("deploying agent %s: %w", name, err)
			result.Duration = time.Since(start)
			return result
		}
	}
	defer e.AgentManager.StopAll(ctx)

	// 2. Apply initial state.
	if err := e.applySetup(ctx, &s.Setup); err != nil {
		result.Error = fmt.Errorf("applying setup: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// 3. Fire trigger.
	if s.Trigger != nil {
		if err := e.applyTrigger(ctx, s.Trigger); err != nil {
			result.Error = fmt.Errorf("applying trigger: %w", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// 4. Poll for expected state.
	if err := e.waitForExpectations(ctx, s.Expect); err != nil {
		result.Error = err
		result.Diagnostics = e.collectDiagnostics(ctx, s)
		result.Duration = time.Since(start)
		return result
	}

	result.Passed = true
	result.Duration = time.Since(start)
	return result
}

func (e *Engine) applySetup(ctx context.Context, setup *scenario.Setup) error {
	for _, manifest := range setup.Manifests {
		if err := e.applyManifest(ctx, manifest); err != nil {
			return fmt.Errorf("applying manifest %s: %w", manifest, err)
		}
	}
	return nil
}

// applyManifest reads a YAML file (potentially multi-document) and creates/updates
// each resource using the dynamic client.
func (e *Engine) applyManifest(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening manifest: %w", err)
	}
	defer f.Close()

	decoder := yamlutil.NewYAMLOrJSONDecoder(f, 4096)
	for {
		var obj unstructured.Unstructured
		err := decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("decoding YAML document: %w", err)
		}
		if obj.Object == nil {
			continue
		}
		if err := e.applyUnstructured(ctx, &obj); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) applyUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()
	mapping, err := e.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("finding REST mapping for %s: %w", gvk, err)
	}

	var resource dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := obj.GetNamespace()
		if ns == "" {
			ns = "default"
		}
		resource = e.dynamicClient.Resource(mapping.Resource).Namespace(ns)
	} else {
		resource = e.dynamicClient.Resource(mapping.Resource)
	}

	_, err = resource.Create(ctx, obj, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		existing, getErr := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("getting existing resource %s/%s: %w", obj.GetKind(), obj.GetName(), getErr)
		}
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = resource.Update(ctx, obj, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("applying resource %s/%s: %w", obj.GetKind(), obj.GetName(), err)
	}
	return nil
}

func (e *Engine) applyTrigger(ctx context.Context, trigger *scenario.Trigger) error {
	if trigger.Patch != nil {
		return e.applyPatch(ctx, trigger.Patch)
	}
	return nil
}

// applyPatch applies a strategic merge patch to the specified resource.
func (e *Engine) applyPatch(ctx context.Context, patch *scenario.ResourcePatch) error {
	gvk := schema.FromAPIVersionAndKind(patch.APIVersion, patch.Kind)
	mapping, err := e.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("finding REST mapping for %s: %w", gvk, err)
	}

	patchData := map[string]interface{}{}
	if patch.Spec != nil {
		patchData["spec"] = patch.Spec
	}
	patchBytes, err := json.Marshal(patchData)
	if err != nil {
		return fmt.Errorf("marshaling patch: %w", err)
	}

	var resource dynamic.ResourceInterface
	ns := patch.Namespace
	if ns == "" {
		ns = "default"
	}
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resource = e.dynamicClient.Resource(mapping.Resource).Namespace(ns)
	} else {
		resource = e.dynamicClient.Resource(mapping.Resource)
	}

	_, err = resource.Patch(ctx, patch.Name, "application/merge-patch+json", patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("patching %s/%s: %w", patch.Kind, patch.Name, err)
	}
	return nil
}

func (e *Engine) waitForExpectations(ctx context.Context, expectations []scenario.Expectation) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Check once immediately before waiting.
	allMet, lastErr := e.checkAllExpectations(ctx, expectations)
	if allMet {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			if lastErr != "" {
				return fmt.Errorf("timed out waiting for expectations (last mismatch: %s): %w", lastErr, ctx.Err())
			}
			return fmt.Errorf("timed out waiting for expectations: %w", ctx.Err())
		case <-ticker.C:
			allMet, lastErr = e.checkAllExpectations(ctx, expectations)
			if allMet {
				return nil
			}
		}
	}
}

func (e *Engine) checkAllExpectations(ctx context.Context, expectations []scenario.Expectation) (bool, string) {
	for _, exp := range expectations {
		met, detail := e.checkExpectation(ctx, &exp)
		if !met {
			return false, detail
		}
	}
	return true, ""
}

// checkExpectation fetches a resource and evaluates all conditions against it.
// Returns true if all conditions are met, plus a detail string on mismatch.
func (e *Engine) checkExpectation(ctx context.Context, exp *scenario.Expectation) (bool, string) {
	ref := exp.Resource
	gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
	mapping, err := e.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return false, fmt.Sprintf("no mapping for %s: %v", gvk, err)
	}

	ns := ref.Namespace
	if ns == "" {
		ns = "default"
	}
	var resource dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resource = e.dynamicClient.Resource(mapping.Resource).Namespace(ns)
	} else {
		resource = e.dynamicClient.Resource(mapping.Resource)
	}

	obj, err := resource.Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Sprintf("getting %s/%s: %v", ref.Kind, ref.Name, err)
	}

	for _, cond := range exp.Conditions {
		actual, found, err := lookupPath(obj.Object, cond.Path)
		if err != nil || !found {
			return false, fmt.Sprintf("%s/%s path %s: not found", ref.Kind, ref.Name, cond.Path)
		}
		if !valuesEqual(actual, cond.Value) {
			return false, fmt.Sprintf("%s/%s path %s: got %v, want %v", ref.Kind, ref.Name, cond.Path, actual, cond.Value)
		}
	}
	return true, ""
}

// lookupPath navigates a nested map using a dot-separated path like ".spec.replicas".
func lookupPath(obj map[string]interface{}, path string) (interface{}, bool, error) {
	path = strings.TrimPrefix(path, ".")
	parts := strings.Split(path, ".")

	var current interface{} = obj
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("expected map at %q, got %T", part, current)
		}
		current, ok = m[part]
		if !ok {
			return nil, false, nil
		}
	}
	return current, true, nil
}

// valuesEqual compares expected and actual values, handling numeric type coercion.
// YAML numbers parse as int, but Kubernetes JSON responses use float64.
func valuesEqual(actual, expected interface{}) bool {
	// Convert both to float64 for numeric comparison.
	af, aOk := toFloat64(actual)
	ef, eOk := toFloat64(expected)
	if aOk && eOk {
		return af == ef
	}
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	}
	return 0, false
}

func (e *Engine) collectDiagnostics(ctx context.Context, s *scenario.Scenario) *diagnostics.Report {
	if e.Collector == nil {
		return nil
	}

	report, err := e.Collector.Collect(ctx, diagnostics.Scope{
		Namespace:  e.inferNamespace(s),
		AgentNames: s.Agents,
	})
	if err != nil {
		return &diagnostics.Report{
			CollectionError: err.Error(),
		}
	}
	return report
}

func (e *Engine) inferNamespace(s *scenario.Scenario) string {
	for _, exp := range s.Expect {
		if exp.Resource.Namespace != "" {
			return exp.Resource.Namespace
		}
	}
	return "default"
}
