package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lazyk8s/lazy-k8s/internal/config"
	"github.com/lazyk8s/lazy-k8s/internal/k8s"
	"github.com/lazyk8s/lazy-k8s/internal/ui"
)

type App struct {
	config    *config.Config
	k8sClient *k8s.Client
}

func New(cfg *config.Config) (*App, error) {
	client, err := k8s.NewClient(cfg.Kubeconfig, cfg.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Set namespace from config if specified
	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	return &App{
		config:    cfg,
		k8sClient: client,
	}, nil
}

func (a *App) Run() error {
	model := ui.NewModel(a.k8sClient, a.config)

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := program.Run()

	return err
}

func (a *App) Client() *k8s.Client {
	return a.k8sClient
}

func (a *App) Config() *config.Config {
	return a.config
}
