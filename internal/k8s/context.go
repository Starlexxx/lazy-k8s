package k8s

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type ContextInfo struct {
	Name      string
	Cluster   string
	AuthInfo  string
	Namespace string
	IsCurrent bool
}

func GetAllContexts(kubeconfig string) ([]ContextInfo, string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	config, err := loadingRules.Load()
	if err != nil {
		return nil, "", err
	}

	contexts := make([]ContextInfo, 0, len(config.Contexts))
	for name, ctx := range config.Contexts {
		contexts = append(contexts, ContextInfo{
			Name:      name,
			Cluster:   ctx.Cluster,
			AuthInfo:  ctx.AuthInfo,
			Namespace: ctx.Namespace,
			IsCurrent: name == config.CurrentContext,
		})
	}

	return contexts, config.CurrentContext, nil
}

func SetCurrentContext(kubeconfig, contextName string) error {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	config, err := loadingRules.Load()
	if err != nil {
		return err
	}

	config.CurrentContext = contextName

	return clientcmd.ModifyConfig(loadingRules, *config, true)
}

func GetContextNamespaces(kubeconfig, contextName string) (string, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	config, err := loadingRules.Load()
	if err != nil {
		return "", err
	}

	if ctx, ok := config.Contexts[contextName]; ok {
		return ctx.Namespace, nil
	}

	return "default", nil
}

func GetRawConfig(kubeconfig string) (*api.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	return loadingRules.Load()
}
