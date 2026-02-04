.PHONY: build test lint clean release-snapshot install help

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Default target
help:
	@echo "FlowCLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build            Build the binary"
	@echo "  make install          Install to GOPATH/bin"
	@echo "  make test             Run tests"
	@echo "  make test-coverage    Run tests with coverage"
	@echo "  make lint             Run linter"
	@echo "  make clean            Remove build artifacts"
	@echo "  make release-snapshot Build snapshot release (for testing)"
	@echo "  make release-dry-run  Test release process without publishing"

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o flowcli ./cmd/flowcli

# Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/flowcli

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f flowcli
	rm -f coverage.out coverage.html
	rm -rf dist/

# Build snapshot release (for local testing)
release-snapshot:
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser release --snapshot --clean

# Test release process without publishing
release-dry-run:
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)
	goreleaser release --skip=publish --clean

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Run all checks (useful before committing)
check: fmt tidy lint test
	@echo "All checks passed!"
