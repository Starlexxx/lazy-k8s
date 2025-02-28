package client

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Test context name constant
const testContextName = "test-context"

// Original functions are commented out as they're not used in tests yet
// but might be useful for more advanced mocking in the future
/*
var (
	origBuildConfigFromFlags = clientcmd.BuildConfigFromFlags
	origNewForConfig         = kubernetes.NewForConfig
	origInClusterConfig      = rest.InClusterConfig
)
*/

func TestNewK8sClient(t *testing.T) {
	// Create a temporary directory for test kubeconfig
	tempDir, err := os.MkdirTemp("", "k8s-client-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test kubeconfig
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")

	// Create test kubeconfig content
	cfg := clientcmdapi.NewConfig()
	cfg.APIVersion = "v1"
	cfg.Kind = "Config"

	// Add cluster
	cfg.Clusters["test-cluster"] = &clientcmdapi.Cluster{
		Server: "https://test-server:6443",
	}

	// Add context
	cfg.Contexts[testContextName] = &clientcmdapi.Context{
		Cluster:  "test-cluster",
		AuthInfo: "test-user",
	}

	// Set current context
	cfg.CurrentContext = testContextName

	// Save kubeconfig
	err = clientcmd.WriteToFile(*cfg, kubeConfigPath)
	if err != nil {
		t.Fatalf("Error writing test kubeconfig: %v", err)
	}

	// Test successful client creation with correct configuration
	t.Run("Successful client creation", func(t *testing.T) {
		client, err := NewK8sClient(kubeConfigPath)
		if err != nil {
			t.Errorf("Error creating client: %v", err)
		}

		if client == nil {
			t.Fatal("Client should not be nil")
		}

		if client.CurrentContext != testContextName {
			t.Errorf("Incorrect current context, got: %s, expected: %s", client.CurrentContext, testContextName)
		}

		if client.ConfigPath != kubeConfigPath {
			t.Errorf("Incorrect config path, got: %s, expected: %s", client.ConfigPath, kubeConfigPath)
		}
	})

	// Test client creation without specifying path
	t.Run("Client creation without specifying path", func(t *testing.T) {
		// Save original HOME environment variable
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)

		// Set HOME to temporary directory
		os.Setenv("HOME", tempDir)

		// Create .kube directory
		kubeDir := filepath.Join(tempDir, ".kube")
		err := os.MkdirAll(kubeDir, 0755)
		if err != nil {
			t.Fatalf("Error creating .kube directory: %v", err)
		}

		// Copy test kubeconfig to ~/.kube/config
		defaultConfig := filepath.Join(kubeDir, "config")
		err = copyFile(kubeConfigPath, defaultConfig)
		if err != nil {
			t.Fatalf("Error copying kubeconfig: %v", err)
		}

		// Test client creation without specifying path
		client, err := NewK8sClient("")
		if err != nil {
			t.Errorf("Error creating client: %v", err)
		}

		if client == nil {
			t.Fatal("Client should not be nil")
		}

		if client.CurrentContext != testContextName {
			t.Errorf("Incorrect current context, got: %s, expected: %s", client.CurrentContext, testContextName)
		}
	})
}

func TestGetInClusterClient(t *testing.T) {
	t.Skip("Skipping in-cluster client test as it requires special environment conditions")
}

// Helper function for copying files
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
