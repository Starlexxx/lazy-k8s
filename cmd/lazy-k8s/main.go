package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Starlexxx/lazy-k8s/internal/app"
	"github.com/Starlexxx/lazy-k8s/internal/config"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "lazy-k8s",
		Short: "A terminal UI for Kubernetes management",
		Long: `lazy-k8s is a terminal-based user interface for managing Kubernetes clusters.
Inspired by lazygit's design philosophy, it provides an intuitive,
keyboard-driven interface for common Kubernetes operations.`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
			context, _ := cmd.Flags().GetString("context")
			namespace, _ := cmd.Flags().GetString("namespace")

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if kubeconfig != "" {
				cfg.Kubeconfig = kubeconfig
			}

			if context != "" {
				cfg.Context = context
			}

			if namespace != "" {
				cfg.Namespace = namespace
			}

			application, err := app.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize app: %w", err)
			}

			return application.Run()
		},
	}

	rootCmd.Flags().StringP("kubeconfig", "k", "", "Path to kubeconfig file")
	rootCmd.Flags().StringP("context", "c", "", "Kubernetes context to use")
	rootCmd.Flags().StringP("namespace", "n", "", "Kubernetes namespace to use")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
