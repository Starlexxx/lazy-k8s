package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Starlexxx/lazy-k8s/pkg/client"
	"github.com/Starlexxx/lazy-k8s/pkg/commands"
	"github.com/Starlexxx/lazy-k8s/pkg/ui"
	"github.com/spf13/cobra"
)

// Build information. Populated at build-time by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var skipK8sCheck bool

	rootCmd := &cobra.Command{
		Use:   "lazy-k8s",
		Short: "lazy-k8s is a simplified kubectl replacement",
		Long:  `A simplified and intuitive CLI tool for managing Kubernetes clusters with a modern UI`,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			// Skip checks for help command
			if cmd.Name() == "help" {
				return
			}
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVar(&skipK8sCheck, "skip-k8s-check", false, "Skip Kubernetes client initialization (for offline testing)")

	// Create Kubernetes client
	var k8sClient *client.K8sClient
	var clientErr error

	// Add version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of lazy-k8s",
		Long:  `All software has versions. This is lazy-k8s's`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("lazy-k8s version %s\n", version)
			fmt.Printf("  Built on %s\n", date)
			fmt.Printf("  Git commit: %s\n", commit)
			fmt.Printf("  Go version: %s\n", runtime.Version())
			fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Add get command
	getCmd := &cobra.Command{
		Use:   "get [resource]",
		Short: "Display one or many resources",
		Long:  `Prints a table of the most important information about the specified resources.`,
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(_ *cobra.Command, _ []string) {
			// Initialize client before executing command
			if k8sClient == nil && !skipK8sCheck {
				k8sClient, clientErr = client.NewK8sClient("")
				if clientErr != nil {
					fmt.Fprintf(os.Stderr, "Error: Error initializing Kubernetes client: %v\nMake sure your kubeconfig is properly configured\n", clientErr)
					os.Exit(1)
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to get namespace flag: %v\n", err)
				os.Exit(1)
			}
			resource := args[0]
			commands.Get(k8sClient, resource, namespace)
		},
	}
	getCmd.Flags().StringP("namespace", "n", "default", "Specify the namespace to use")
	rootCmd.AddCommand(getCmd)

	// Add UI command
	uiCmd := &cobra.Command{
		Use:   "ui",
		Short: "Start the interactive terminal UI",
		Long:  `Launch an interactive terminal UI for managing Kubernetes resources`,
		PreRun: func(_ *cobra.Command, _ []string) {
			// Initialize client before executing command
			if !skipK8sCheck {
				k8sClient, clientErr = client.NewK8sClient("")
				if clientErr != nil {
					fmt.Fprintf(os.Stderr, "Error: Error initializing Kubernetes client: %v\nMake sure your kubeconfig is properly configured\n", clientErr)
					os.Exit(1)
				}
			}
		},
		Run: func(_ *cobra.Command, _ []string) {
			app := ui.NewApp(k8sClient)
			if err := app.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(uiCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
