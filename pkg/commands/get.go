package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/Starlexxx/lazy-k8s/pkg/client"
	"github.com/Starlexxx/lazy-k8s/pkg/utils"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Get is a simple wrapper function to get resources for direct use
func Get(k8sClient *client.K8sClient, resource string, namespace string) {
	switch resource {
	case "pods", "pod", "po":
		if err := getPods(k8sClient, namespace, "", false); err != nil {
			fmt.Printf("Error getting pods: %v\n", err)
		}
	case "nodes", "node", "no":
		if err := getNodes(k8sClient, "", false); err != nil {
			fmt.Printf("Error getting nodes: %v\n", err)
		}
	case "services", "service", "svc":
		fmt.Println("Getting services information - functionality will be implemented later")
	case "deployments", "deployment", "deploy":
		fmt.Println("Getting deployments information - functionality will be implemented later")
	default:
		fmt.Printf("Unknown resource type: %s\n", resource)
	}
}

// displayPodDetails shows detailed information about a specific pod
func displayPodDetails(pod *corev1.Pod, showLabels bool) error {
	fmt.Printf("Name: %s\n", pod.Name)
	fmt.Printf("Namespace: %s\n", pod.Namespace)
	fmt.Printf("Status: %s\n", pod.Status.Phase)
	fmt.Printf("IP: %s\n", pod.Status.PodIP)
	fmt.Printf("Created: %v (Age: %s)\n", pod.CreationTimestamp.Format(time.RFC3339), utils.FormatAge(pod.CreationTimestamp.Time))

	if showLabels && len(pod.Labels) > 0 {
		fmt.Println("Labels:")
		for key, value := range pod.Labels {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("Containers:")
	for _, container := range pod.Spec.Containers {
		fmt.Printf("  - %s (Image: %s)\n", container.Name, container.Image)
	}

	return nil
}

// displayNodeDetails shows detailed information about a specific node
func displayNodeDetails(node *corev1.Node, showLabels bool) error {
	fmt.Printf("Name: %s\n", node.Name)
	fmt.Printf("Status: %v\n", utils.GetNodeConditionStatus(node.Status.Conditions))
	fmt.Printf("Roles: %s\n", utils.GetNodeRoles(node.Labels))
	fmt.Printf("Kubernetes Version: %s\n", node.Status.NodeInfo.KubeletVersion)
	fmt.Printf("OS: %s\n", node.Status.NodeInfo.OSImage)
	fmt.Printf("Architecture: %s\n", node.Status.NodeInfo.Architecture)
	fmt.Printf("Container Runtime: %s\n", node.Status.NodeInfo.ContainerRuntimeVersion)
	fmt.Printf("Created: %v (Age: %s)\n", node.CreationTimestamp.Format(time.RFC3339), utils.FormatAge(node.CreationTimestamp.Time))

	if showLabels && len(node.Labels) > 0 {
		fmt.Println("Labels:")
		for key, value := range node.Labels {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

// getPods gets information about pods
func getPods(k8sClient *client.K8sClient, namespace string, podName string, showLabels bool) error {
	if podName != "" {
		pod, err := k8sClient.ClientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting pod %s info: %w", podName, err)
		}

		return displayPodDetails(pod, showLabels)
	}

	// Get list of pods
	pods, err := k8sClient.ClientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting pods list: %w", err)
	}

	if len(pods.Items) == 0 {
		fmt.Printf("No pods found in namespace %s\n", namespace)
		return nil
	}

	// Output in table format
	fmt.Printf("NAMESPACE\tNAME\tSTATUS\tIP\tAGE")
	if showLabels {
		fmt.Printf("\tLABELS")
	}
	fmt.Println()

	for _, pod := range pods.Items {
		age := utils.FormatAge(pod.CreationTimestamp.Time)
		fmt.Printf("%s\t%s\t%s\t%s\t%s", pod.Namespace, pod.Name, pod.Status.Phase, pod.Status.PodIP, age)

		if showLabels {
			labels := utils.FormatLabels(pod.Labels)
			fmt.Printf("\t%s", labels)
		}

		fmt.Println()
	}

	return nil
}

// getNodes gets information about nodes
func getNodes(k8sClient *client.K8sClient, nodeName string, showLabels bool) error {
	if nodeName != "" {
		node, err := k8sClient.ClientSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error getting node %s info: %w", nodeName, err)
		}

		return displayNodeDetails(node, showLabels)
	}

	// Get list of nodes
	nodes, err := k8sClient.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error getting nodes list: %w", err)
	}

	if len(nodes.Items) == 0 {
		fmt.Println("No nodes found in the cluster")
		return nil
	}

	// Output in table format
	fmt.Printf("NAME\tSTATUS\tROLES\tAGE\tVERSION")
	if showLabels {
		fmt.Printf("\tLABELS")
	}
	fmt.Println()

	for _, node := range nodes.Items {
		age := utils.FormatAge(node.CreationTimestamp.Time)
		status := utils.GetNodeConditionStatus(node.Status.Conditions)
		roles := utils.GetNodeRoles(node.Labels)
		version := node.Status.NodeInfo.KubeletVersion

		fmt.Printf("%s\t%s\t%s\t%s\t%s", node.Name, status, roles, age, version)

		if showLabels {
			labels := utils.FormatLabels(node.Labels)
			fmt.Printf("\t%s", labels)
		}

		fmt.Println()
	}

	return nil
}

// NewGetCommand creates 'get' command and adds subcommands to it
func NewGetCommand(k8sClient *client.K8sClient) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get information about Kubernetes resources",
		Long:  "Get information about various Kubernetes resources such as pods, services, nodes, etc.",
		Run: func(cmd *cobra.Command, _ []string) {
			// If no subcommands specified, show help
			if err := cmd.Help(); err != nil {
				fmt.Printf("Error displaying help: %v\n", err)
			}
		},
	}

	// Add subcommands
	getCmd.AddCommand(
		newGetPodsCommand(k8sClient),
		newGetNodesCommand(k8sClient),
		newGetServicesCommand(k8sClient),
		newGetDeploymentsCommand(k8sClient),
	)

	return getCmd
}

// newGetPodsCommand creates a command for getting pod information
func newGetPodsCommand(k8sClient *client.K8sClient) *cobra.Command {
	var namespace string
	var showLabels bool

	podsCmd := &cobra.Command{
		Use:   "pods [NAME]",
		Short: "Get information about pods",
		Long:  "Get a list of pods in the specified namespace or information about a specific pod",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			podName := ""
			if len(args) > 0 {
				podName = args[0]
			}

			return getPods(k8sClient, namespace, podName, showLabels)
		},
	}

	podsCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to get pods from")
	podsCmd.Flags().BoolVarP(&showLabels, "show-labels", "l", false, "Show pod labels")

	return podsCmd
}

// newGetNodesCommand creates a command for getting node information
func newGetNodesCommand(k8sClient *client.K8sClient) *cobra.Command {
	var showLabels bool

	nodesCmd := &cobra.Command{
		Use:   "nodes [NAME]",
		Short: "Get information about nodes",
		Long:  "Get a list of nodes in the cluster or information about a specific node",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			nodeName := ""
			if len(args) > 0 {
				nodeName = args[0]
			}

			return getNodes(k8sClient, nodeName, showLabels)
		},
	}

	nodesCmd.Flags().BoolVarP(&showLabels, "show-labels", "l", false, "Show node labels")

	return nodesCmd
}

// newGetServicesCommand creates a command for getting service information
func newGetServicesCommand(_ *client.K8sClient) *cobra.Command {
	return &cobra.Command{
		Use:   "services",
		Short: "Get information about services",
		Long:  "Get a list of services in the specified namespace or information about a specific service",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("Getting services information - functionality will be implemented later")
		},
	}
}

// newGetDeploymentsCommand creates a command for getting deployment information
func newGetDeploymentsCommand(_ *client.K8sClient) *cobra.Command {
	return &cobra.Command{
		Use:   "deployments",
		Short: "Get information about deployments",
		Long:  "Get a list of deployments in the specified namespace or information about a specific deployment",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("Getting deployments information - functionality will be implemented later")
		},
	}
}
