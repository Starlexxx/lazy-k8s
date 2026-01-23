# lazy-k8s

A terminal-based user interface for Kubernetes management, inspired by lazygit.

## Features

- **Multi-panel layout** with keyboard-driven navigation
- **Real-time updates** using Kubernetes watch API
- **Resource management** for:
  - Namespaces
  - Pods (logs, exec, delete)
  - Deployments (scale, restart, rollback)
  - Services (port-forward)
  - ConfigMaps
  - Secrets
  - Nodes
  - Events
- **Context and namespace switching**
- **Search/filter** within panels
- **YAML viewer** with syntax highlighting
- **Log streaming** with follow mode

## Installation

### From Source

```bash
git clone https://github.com/lazyk8s/lazy-k8s.git
cd lazy-k8s
make build
./bin/lazy-k8s
```

### Go Install

```bash
go install github.com/lazyk8s/lazy-k8s/cmd/lazy-k8s@latest
```

## Usage

```bash
# Use default kubeconfig
lazy-k8s

# Specify kubeconfig
lazy-k8s -k ~/.kube/config

# Use specific context
lazy-k8s -c my-cluster

# Start in specific namespace
lazy-k8s -n kube-system
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `g` | Go to top |
| `G` | Go to bottom |
| `Tab` | Next panel |
| `Shift+Tab` | Previous panel |
| `1-9` | Jump to panel |
| `Enter` | Select/expand |
| `Esc` | Back/cancel |

### General

| Key | Action |
|-----|--------|
| `?` | Show help |
| `q` / `Ctrl+c` | Quit |
| `/` | Search/filter |
| `Ctrl+r` | Refresh |
| `c` | Switch context |
| `n` | Switch namespace |
| `A` | Toggle all namespaces |

### Resource Actions

| Key | Action |
|-----|--------|
| `d` | Describe resource |
| `y` | View YAML |
| `e` | Edit resource |
| `D` | Delete (with confirm) |
| `Ctrl+y` | Copy YAML |

### Pod Actions

| Key | Action |
|-----|--------|
| `l` | View logs |
| `f` | Toggle follow logs |
| `x` | Exec into container |
| `p` | Port forward |

### Deployment Actions

| Key | Action |
|-----|--------|
| `s` | Scale |
| `r` | Restart (rollout) |
| `R` | Rollback |

## Configuration

Configuration file location: `~/.config/lazy-k8s/config.yaml`

```yaml
theme:
  primaryColor: "#7aa2f7"
  secondaryColor: "#9ece6a"
  errorColor: "#f7768e"
  warningColor: "#e0af68"
  backgroundColor: "#1a1b26"
  textColor: "#c0caf5"
  borderColor: "#3b4261"

defaults:
  namespace: "default"
  logLines: 100
  followLogs: true
  refreshInterval: 5

panels:
  visible:
    - namespaces
    - pods
    - deployments
    - services
```

## Requirements

- Go 1.21+
- kubectl configured with cluster access
- Terminal with 256 color support

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linter
make lint
```

## License

MIT License
