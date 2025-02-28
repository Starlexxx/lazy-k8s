# lazy-k8s

[![CI](https://github.com/Starlexxx/lazy-k8s/actions/workflows/ci.yml/badge.svg)](https://github.com/Starlexxx/lazy-k8s/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/Starlexxx/lazy-k8s/branch/main/graph/badge.svg)](https://codecov.io/gh/Starlexxx/lazy-k8s)
[![Go Report Card](https://goreportcard.com/badge/github.com/Starlexxx/lazy-k8s)](https://goreportcard.com/report/github.com/Starlexxx/lazy-k8s)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Starlexxx/lazy-k8s)](https://github.com/Starlexxx/lazy-k8s)
[![License](https://img.shields.io/github/license/Starlexxx/lazy-k8s)](https://github.com/Starlexxx/lazy-k8s/blob/main/LICENSE)

A command-line utility that simplifies interaction with Kubernetes. It provides convenient commands for managing clusters, pods, services, and other resources. Includes both CLI commands and an interactive terminal user interface (TUI).

## Features

### Implemented Features

- Get general information about the cluster: node status, pods, services, etc.
- Interactive terminal user interface (TUI) similar to lazygit for convenient viewing and management of Kubernetes resources
- Navigation between panels in TUI to view pods, nodes, and other information
- Detailed view of pod and node information
- Formatted output of age, labels, node condition status, and roles

### Planned Features

- Manage pods: create, delete, scale, get logs, etc.
- Work with configurations: create and edit ConfigMap and Secret.
- Manage deployments: create, update, rollback.
- Monitor resources: view CPU, memory, and other resource usage by pods and nodes.
- Search for objects by name or labels.
- Execute commands in pods (e.g., exec).
- Support for working with multiple clusters and contexts.
- Automatic completion of commands and object names for ease of use.
- Ability to save frequently used commands and configurations for reuse.

## Installation

### Prerequisites

- Go 1.21 or higher
- A configured Kubernetes cluster (and a valid kubeconfig file)

### From Release

Download the pre-built binary for your platform from the [releases page](https://github.com/Starlexxx/lazy-k8s/releases).

```bash
# Download the latest release (adjust for your platform)
curl -LO https://github.com/Starlexxx/lazy-k8s/releases/latest/download/lazy-k8s_Linux_x86_64.tar.gz

# Extract the binary
tar -xzf lazy-k8s_Linux_x86_64.tar.gz

# Move to a directory in your PATH
sudo mv lazy-k8s /usr/local/bin/
```

### From Source

1. Clone the repository:

```bash
git clone https://github.com/Starlexxx/lazy-k8s.git
cd lazy-k8s
```

2. Build the project:

```bash
go build -o lazy-k8s ./cmd/lazy-k8s
```

3. (Optional) Move the binary to a directory in your PATH:

```bash
mv lazy-k8s /usr/local/bin/
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/). For the versions available, see the [tags on this repository](https://github.com/Starlexxx/lazy-k8s/tags).

To check the current version of your lazy-k8s installation:

```bash
lazy-k8s version
```

## Usage

### Basic Commands

```bash
# Get help
lazy-k8s help

# Display version information
lazy-k8s version

# Get pods in the default namespace
lazy-k8s get pods

# Get pods in a specific namespace
lazy-k8s get pods -n kube-system

# Get detailed information about a specific pod
lazy-k8s get pods my-pod-name

# Get nodes in the cluster
lazy-k8s get nodes

# Get detailed information about a specific node
lazy-k8s get nodes my-node-name

# Launch the interactive terminal user interface
lazy-k8s ui
```

### Using the Terminal Interface (TUI)

The terminal user interface provides a convenient way to view and manage Kubernetes resources:

- Use arrow keys to navigate between panels and elements
- Press `Tab` to switch between main panels
- Press `Enter` to select and view detailed information about a resource
- Press `Esc` to return or close modal windows
- Press `Ctrl+C` to exit the application

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
