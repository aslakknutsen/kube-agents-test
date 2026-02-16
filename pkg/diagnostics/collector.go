package diagnostics

import (
	"bytes"
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterCollector gathers diagnostics from a live Kubernetes cluster.
type ClusterCollector struct {
	KubeconfigPath string

	client kubernetes.Interface
}

// NewClusterCollector creates a collector that connects via the given kubeconfig.
func NewClusterCollector(kubeconfigPath string) *ClusterCollector {
	return &ClusterCollector{KubeconfigPath: kubeconfigPath}
}

// NewClusterCollectorWithClient creates a collector using an injected client (for testing).
func NewClusterCollectorWithClient(client kubernetes.Interface) *ClusterCollector {
	return &ClusterCollector{client: client}
}

func (c *ClusterCollector) getClient() (kubernetes.Interface, error) {
	if c.client != nil {
		return c.client, nil
	}
	config, err := clientcmd.BuildConfigFromFlags("", c.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("building kubeconfig: %w", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}
	c.client = client
	return client, nil
}

// Collect gathers agent logs and Kubernetes events for the given scope.
func (c *ClusterCollector) Collect(ctx context.Context, scope Scope) (*Report, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	report := &Report{
		AgentLogs: make(map[string]string),
	}

	// Collect agent logs.
	for _, agentName := range scope.AgentNames {
		logs, err := c.collectAgentLogs(ctx, client, scope.Namespace, agentName)
		if err != nil {
			report.AgentLogs[agentName] = fmt.Sprintf("error collecting logs: %v", err)
		} else {
			report.AgentLogs[agentName] = logs
		}
	}

	// Collect events.
	events, err := client.CoreV1().Events(scope.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		report.CollectionError = fmt.Sprintf("listing events: %v", err)
	} else {
		for _, ev := range events.Items {
			report.Events = append(report.Events, fmt.Sprintf(
				"%s %s/%s: %s (%s)",
				ev.LastTimestamp.Format("15:04:05"),
				ev.InvolvedObject.Kind,
				ev.InvolvedObject.Name,
				ev.Message,
				ev.Reason,
			))
		}
	}

	return report, nil
}

func (c *ClusterCollector) collectAgentLogs(ctx context.Context, client kubernetes.Interface, namespace, agentName string) (string, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", agentName),
	})
	if err != nil {
		return "", fmt.Errorf("listing pods: %w", err)
	}

	var buf bytes.Buffer
	for _, pod := range pods.Items {
		stream, err := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			TailLines: int64Ptr(200),
		}).Stream(ctx)
		if err != nil {
			fmt.Fprintf(&buf, "--- pod %s: error: %v\n", pod.Name, err)
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
