package client

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// K8sClient is a wrapper around various Kubernetes clients
type K8sClient struct {
	// ClientSet for interacting with main Kubernetes resources
	ClientSet kubernetes.Interface
	// MetricsClient for fetching metrics
	MetricsClient metricsv.Interface
	// CurrentContext of the configuration
	CurrentContext string
	// ConfigPath to the kubeconfig file
	ConfigPath string
}

// NewK8sClient creates a new K8sClient using the kubeconfig configuration
func NewK8sClient(configPath string) (*K8sClient, error) {
	// If config path is not specified, use the default path
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(home, ".kube", "config")
	}

	// Load the configuration
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	// Create a client for the Kubernetes API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create a client for the metrics API
	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Get the current context
	kubeConfig, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		ClientSet:      clientset,
		MetricsClient:  metricsClient,
		CurrentContext: kubeConfig.CurrentContext,
		ConfigPath:     configPath,
	}, nil
}

// GetInClusterClient creates a client when running inside a Kubernetes cluster
func GetInClusterClient() (*K8sClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8sClient{
		ClientSet:     clientset,
		MetricsClient: metricsClient,
	}, nil
}

// SwitchContext switches the current context in the configuration
func (c *K8sClient) SwitchContext(context string) error {
	kubeConfig, err := clientcmd.LoadFromFile(c.ConfigPath)
	if err != nil {
		return err
	}

	// Check if the specified context exists
	if _, exists := kubeConfig.Contexts[context]; !exists {
		return err
	}

	// Set the new current context
	kubeConfig.CurrentContext = context

	// Save the changes
	if err := clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), *kubeConfig, true); err != nil {
		return err
	}

	// Update the client with the new context
	newClient, err := NewK8sClient(c.ConfigPath)
	if err != nil {
		return err
	}

	*c = *newClient
	return nil
}
