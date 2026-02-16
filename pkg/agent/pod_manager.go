package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// AgentConfig holds the configuration for deploying a single agent.
type AgentConfig struct {
	// Name is the agent identifier (must match the name used in scenario YAML).
	Name string

	// Image is the container image to deploy (e.g. "ghcr.io/kube-agents/scaling-agent:latest").
	Image string

	// Namespace is the Kubernetes namespace to deploy into.
	Namespace string

	// Args are additional command-line arguments passed to the agent container.
	Args []string
}

// PodManager deploys agents as Kubernetes Deployments in the test cluster.
type PodManager struct {
	KubeconfigPath string
	Agents         map[string]AgentConfig

	mu       sync.Mutex
	deployed map[string]bool
	client   kubernetes.Interface
}

// NewPodManager creates a PodManager with the given kubeconfig and agent configs.
func NewPodManager(kubeconfigPath string, agents []AgentConfig) *PodManager {
	m := &PodManager{
		KubeconfigPath: kubeconfigPath,
		Agents:         make(map[string]AgentConfig),
		deployed:       make(map[string]bool),
	}
	for _, a := range agents {
		m.Agents[a.Name] = a
	}
	return m
}

// NewPodManagerWithClient creates a PodManager using an injected client (for testing).
func NewPodManagerWithClient(client kubernetes.Interface, agents []AgentConfig) *PodManager {
	m := &PodManager{
		Agents:   make(map[string]AgentConfig),
		deployed: make(map[string]bool),
		client:   client,
	}
	for _, a := range agents {
		m.Agents[a.Name] = a
	}
	return m
}

func (m *PodManager) getClient() (kubernetes.Interface, error) {
	if m.client != nil {
		return m.client, nil
	}
	config, err := clientcmd.BuildConfigFromFlags("", m.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubeconfig: %w", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}
	m.client = client
	return client, nil
}

func (m *PodManager) Deploy(ctx context.Context, agentName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, ok := m.Agents[agentName]
	if !ok {
		return fmt.Errorf("unknown agent %q", agentName)
	}

	client, err := m.getClient()
	if err != nil {
		return err
	}

	ns := cfg.Namespace
	if ns == "" {
		ns = "default"
	}

	// Ensure namespace exists.
	_, err = client.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err != nil {
		_, createErr := client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		}, metav1.CreateOptions{})
		if createErr != nil {
			return fmt.Errorf("creating namespace %s: %w", ns, createErr)
		}
	}

	labels := map[string]string{
		"app":                          agentName,
		"kube-agents-test/managed-by":  "test-framework",
	}

	var replicas int32 = 1
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agentName,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  agentName,
							Image: cfg.Image,
							Args:  cfg.Args,
						},
					},
				},
			},
		},
	}

	_, err = client.AppsV1().Deployments(ns).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating deployment for agent %s: %w", agentName, err)
	}

	m.deployed[agentName] = true
	return nil
}

func (m *PodManager) Stop(ctx context.Context, agentName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.deployed[agentName] {
		return fmt.Errorf("agent %q is not deployed", agentName)
	}

	cfg, ok := m.Agents[agentName]
	if !ok {
		return fmt.Errorf("unknown agent %q", agentName)
	}

	client, err := m.getClient()
	if err != nil {
		return err
	}

	ns := cfg.Namespace
	if ns == "" {
		ns = "default"
	}

	propagation := metav1.DeletePropagationForeground
	err = client.AppsV1().Deployments(ns).Delete(ctx, agentName, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil {
		return fmt.Errorf("deleting deployment for agent %s: %w", agentName, err)
	}

	delete(m.deployed, agentName)
	return nil
}

func (m *PodManager) StopAll(ctx context.Context) error {
	m.mu.Lock()
	names := make([]string, 0, len(m.deployed))
	for name := range m.deployed {
		names = append(names, name)
	}
	m.mu.Unlock()

	for _, name := range names {
		if err := m.Stop(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (m *PodManager) Logs(ctx context.Context, agentName string) (string, error) {
	cfg, ok := m.Agents[agentName]
	if !ok {
		return "", fmt.Errorf("unknown agent %q", agentName)
	}

	client, err := m.getClient()
	if err != nil {
		return "", err
	}

	ns := cfg.Namespace
	if ns == "" {
		ns = "default"
	}

	// Find pods by label.
	pods, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s,kube-agents-test/managed-by=test-framework", agentName),
	})
	if err != nil {
		return "", fmt.Errorf("listing pods for agent %s: %w", agentName, err)
	}

	var buf bytes.Buffer
	since := time.Hour
	for _, pod := range pods.Items {
		stream, err := client.CoreV1().Pods(ns).GetLogs(pod.Name, &corev1.PodLogOptions{
			SinceSeconds: int64Ptr(int64(since.Seconds())),
		}).Stream(ctx)
		if err != nil {
			fmt.Fprintf(&buf, "--- pod %s: error getting logs: %v\n", pod.Name, err)
			continue
		}
		fmt.Fprintf(&buf, "--- pod %s ---\n", pod.Name)
		io.Copy(&buf, stream)
		stream.Close()
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func int64Ptr(i int64) *int64 { return &i }
