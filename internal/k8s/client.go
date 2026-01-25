package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	clientset   kubernetes.Interface
	config      clientcmd.ClientConfig
	restConfig  *rest.Config
	rawConfig   api.Config
	contextName string
	namespace   string
}

func NewClient(kubeconfig, contextName string) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		configOverrides.CurrentContext = contextName
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	rawConfig, err := config.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw config: %w", err)
	}

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, _, _ := config.Namespace()
	if namespace == "" {
		namespace = "default"
	}

	currentContext := rawConfig.CurrentContext
	if contextName != "" {
		currentContext = contextName
	}

	return &Client{
		clientset:   clientset,
		config:      config,
		restConfig:  restConfig,
		rawConfig:   rawConfig,
		contextName: currentContext,
		namespace:   namespace,
	}, nil
}

func (c *Client) SetNamespace(ns string) {
	c.namespace = ns
}

func (c *Client) CurrentNamespace() string {
	return c.namespace
}

func (c *Client) CurrentContext() string {
	return c.contextName
}

func (c *Client) Clientset() kubernetes.Interface {
	return c.clientset
}

func (c *Client) RestConfig() *rest.Config {
	return c.restConfig
}

func (c *Client) GetContexts() []string {
	contexts := make([]string, 0, len(c.rawConfig.Contexts))
	for name := range c.rawConfig.Contexts {
		contexts = append(contexts, name)
	}

	return contexts
}

func (c *Client) SwitchContext(contextName string) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, _, _ := config.Namespace()
	if namespace == "" {
		namespace = "default"
	}

	c.clientset = clientset
	c.config = config
	c.restConfig = restConfig
	c.contextName = contextName
	c.namespace = namespace

	return nil
}

func (c *Client) Context() context.Context {
	return context.Background()
}
