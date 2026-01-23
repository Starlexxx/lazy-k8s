BINARY_NAME=lazy-k8s
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

GO=go
GOFLAGS=-trimpath

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	${GO} build ${GOFLAGS} ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/lazy-k8s

.PHONY: build-linux
build-linux:
	@echo "Building ${BINARY_NAME} for Linux..."
	GOOS=linux GOARCH=amd64 ${GO} build ${GOFLAGS} ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/lazy-k8s

.PHONY: build-darwin
build-darwin:
	@echo "Building ${BINARY_NAME} for macOS..."
	GOOS=darwin GOARCH=amd64 ${GO} build ${GOFLAGS} ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd/lazy-k8s
	GOOS=darwin GOARCH=arm64 ${GO} build ${GOFLAGS} ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ./cmd/lazy-k8s

.PHONY: build-windows
build-windows:
	@echo "Building ${BINARY_NAME} for Windows..."
	GOOS=windows GOARCH=amd64 ${GO} build ${GOFLAGS} ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/lazy-k8s

.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME}..."
	cp bin/${BINARY_NAME} ${GOPATH}/bin/${BINARY_NAME}

.PHONY: run
run:
	@echo "Running ${BINARY_NAME}..."
	${GO} run ./cmd/lazy-k8s

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

.PHONY: test
test:
	@echo "Running tests..."
	${GO} test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	${GO} test -v -coverprofile=coverage.out ./...
	${GO} tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	${GO} fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	${GO} vet ./...

.PHONY: tidy
tidy:
	@echo "Tidying modules..."
	${GO} mod tidy

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	${GO} mod download

.PHONY: verify
verify: fmt vet lint test

.PHONY: release
release:
	@echo "Creating release..."
	goreleaser release --rm-dist

.PHONY: snapshot
snapshot:
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --rm-dist

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for all platforms"
	@echo "  install      - Install to GOPATH/bin"
	@echo "  run          - Run the application"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  tidy         - Tidy go modules"
	@echo "  deps         - Download dependencies"
	@echo "  verify       - Run all checks (fmt, vet, lint, test)"
	@echo "  release      - Create a release with goreleaser"
	@echo "  help         - Show this help message"
