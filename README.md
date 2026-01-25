# lazy-k8s

[![Go Version](https://img.shields.io/github/go-mod/go-version/Starlexxx/lazy-k8s)](https://go.dev/)
[![CI](https://github.com/Starlexxx/lazy-k8s/actions/workflows/pr.yml/badge.svg)](https://github.com/Starlexxx/lazy-k8s/actions/workflows/pr.yml)
[![Release](https://github.com/Starlexxx/lazy-k8s/actions/workflows/release.yml/badge.svg)](https://github.com/Starlexxx/lazy-k8s/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Starlexxx/lazy-k8s)](https://goreportcard.com/report/github.com/Starlexxx/lazy-k8s)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A terminal-based user interface for Kubernetes management, inspired by [lazygit](https://github.com/jesseduffield/lazygit).

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
task build
./bin/lazy-k8s
```

### Go Install

```bash
go install github.com/Starlexxx/lazy-k8s/cmd/lazy-k8s@latest
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

- Go 1.25+
- kubectl configured with cluster access
- Terminal with 256 color support
- [Task](https://taskfile.dev/) (optional, for development)

## Building

```bash
# Build for current platform
task build

# Build for all platforms
task build:all

# Run the application
task run

# Run tests
task test

# Run tests with coverage
task test:coverage

# Run linter
task lint

# Run all checks (fmt, vet, lint, test)
task verify

# Show all available tasks
task --list
```

## License

MIT License
