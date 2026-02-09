ontributing to lazy-k8s

Thanks for your interest in contributing to lazy-k8s! Whether it’s a bug report, feature request, or a pull request, your help is welcome.

## Getting Started

1. Fork the repository
1. Clone your fork:

```shell
git clone https://github.com/<your-username>/lazy-k8s.git
cd lazy-k8s
```

1. Install dependencies:

- [Go 1.25+](https://go.dev/dl/)
- [Task](https://taskfile.dev/) (task runner)
- Access to a Kubernetes cluster (local options: [minikube](https://minikube.sigs.k8s.io/), [kind](https://kind.sigs.k8s.io/), [k3d](https://k3d.io/))

1. Verify everything works:

```shell
task verify
```

## Development Workflow

### Build and Run

```shell
# Build
task build

# Run
task run

# Or directly
go run ./cmd/lazy-k8s
```

### Testing

```shell
# Run tests
task test

# Run tests with coverage
task test:coverage
```

### Linting

```shell
# Run linter
task lint
```

### Full Check

```shell
# Run fmt, vet, lint, and test in one go
task verify
```

## Submitting Changes

### Bug Reports

Open an [issue](https://github.com/Starlexxx/lazy-k8s/issues/new) with:

- A clear title and description
- Steps to reproduce
- Expected vs actual behavior
- Your environment: OS, Go version, Kubernetes version, terminal emulator

### Feature Requests

Open an [issue](https://github.com/Starlexxx/lazy-k8s/issues/new) describing:

- The problem you’re trying to solve
- Your proposed solution
- Any alternatives you’ve considered

### Pull Requests

1. Create a branch from `master`:

```shell
git checkout -b feature/my-change
```

1. Make your changes. Keep commits focused and atomic.
1. Make sure all checks pass:

```shell
task verify
```

1. Push and open a PR against `master`:

```shell
git push origin feature/my-change
```

1. In your PR description, explain **what** changed and **why**.

### PR Guidelines

- Keep PRs small and focused on a single change
- Follow existing code style and patterns
- Add tests for new functionality
- Update documentation if behavior changes
- Make sure CI passes before requesting review

## Code Style

- Follow standard Go conventions and [Effective Go](https://go.dev/doc/effective_go)
- Run `gofmt` and `golangci-lint` before committing
- Use meaningful variable and function names
- Add comments for exported functions and non-obvious logic

## Project Structure

```
lazy-k8s/
├── cmd/lazy-k8s/     # Application entrypoint
├── internal/          # Internal packages (not importable by other projects)
├── configs/           # Default configuration files
├── assets/            # Images and demo GIFs
├── .github/workflows/ # CI/CD pipelines
├── Taskfile.yml       # Task runner definitions
└── .goreleaser.yaml   # Release configuration
```

## Need Help?

If you’re unsure about anything, open an issue and ask. No question is too small. We’d rather help you contribute than have you struggle in silence.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
