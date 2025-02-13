# Development Guide

First, install nix-foundry following the [installation guide](GETTING-STARTED.md#installation).

## Quick Setup

```bash
# Initialize with development packages
nix-foundry init --auto \
  --shell zsh \
  --editor nvim

# Add development tools
nix-foundry packages add \
  go \
  gopls \
  delve \
  golangci-lint
```

## Core Components

```go
nix-foundry/
├── cmd/          # CLI commands
├── internal/     # Core logic
├── pkg/          # Public APIs
└── test/         # Test suites
```

## Development Tasks

### Testing
```bash
# Run test suite
go test ./...

# Test specific component
go test ./internal/config/...

# Watch mode
go test -watch ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run

# Check all
go vet ./...
```

## Build Process

### Local Build
```bash
# Build binary
go build -o nix-foundry ./cmd/nix-foundry

# Install locally
go install ./cmd/nix-foundry
```

### Release Build
```bash
# Create release
git tag v1.2.3
git push origin v1.2.3

# Verify release
git checkout v1.2.3
go test ./...
```

## Best Practices

See our comprehensive [Best Practices Guide](BEST-PRACTICES.md#development).

Need help? See:
- [Contributing Guide](CONTRIBUTING.md)
- [FAQ](FAQ.md)
- Join [Discord](https://discord.gg/nix-foundry)
