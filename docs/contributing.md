# Contributing to Nix Foundry

Thank you for your interest in contributing to Nix Foundry! This guide will help you get started.

## Development Setup

1. Fork and clone the repository:

```bash
git clone https://github.com/yourusername/nix-foundry.git
cd nix-foundry
```

2. Install development dependencies:

```bash
# Install Go
nix-env -iA nixpkgs.go

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

3. Set up pre-commit hooks:

```bash
pre-commit install
```

## Project Structure

```shell
nix-foundry/
├── cmd/                    # CLI commands
├── docs/                   # Documentation
├── pkg/                    # Core packages
│   ├── constants/         # Shared constants
│   ├── filesystem/        # Filesystem operations
│   ├── nix/              # Nix package management
│   ├── platform/         # Platform-specific code
│   ├── schema/           # Configuration schemas
│   ├── shell/           # Shell integration
│   ├── testing/         # Test utilities
│   └── validator/       # Configuration validation
├── scripts/              # Development scripts
└── service/             # Service implementations
```

## Development Workflow

1. Create a new branch:

```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following our coding standards:

- Use meaningful variable and function names
- Add docstrings to exported functions and types
- Follow Go best practices and idioms

3. Run tests:

```bash
go test ./...
```

4. Run linter:

```bash
golangci-lint run
```

5. Commit your changes:

```bash
git add .
git commit -m "feat: add new feature"
```

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or modifying tests
- `build`: Build system changes
- `ci`: CI configuration changes
- `chore`: Other changes

## Pull Request Process

1. Update documentation:
- Add/update docstrings
- Update relevant documentation files
- Add examples if applicable

2. Run all tests:
```bash
go test -v -race ./...
```

3. Push your changes:
```bash
git push origin feature/your-feature-name
```

4. Create a Pull Request:
- Use a clear title following conventional commits
- Describe your changes in detail
- Reference any related issues
- Add screenshots if applicable

## Testing Guidelines

1. Write tests for new features:
```go
func TestNewFeature(t *testing.T) {
    // Arrange
    input := "test"

    // Act
    result := NewFeature(input)

    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

2. Include platform-specific tests:
```go
func TestPlatformSpecific(t *testing.T) {
    if runtime.GOOS != "darwin" {
        t.Skip("skipping macOS-specific test")
    }
    // Test macOS-specific functionality
}
```

3. Use test utilities:
```go
import "github.com/shawnkhoffman/nix-foundry/pkg/testing"

func TestWithUtilities(t *testing.T) {
    result := testing.RunPlatformTest(t, func(pt *testing.PlatformTest) {
        // Test implementation
    })
}
```

## Documentation Guidelines

1. Package documentation:
```go
// Package example provides functionality for X.
package example

// ExampleFunc demonstrates how to use X.
func ExampleFunc() {
    // Implementation
}
```

2. Include examples:
```go
func ExampleExampleFunc() {
    result := ExampleFunc()
    fmt.Println(result)
    // Output: expected result
}
```

## Release Process

1. Releases are automated through GitHub Actions:
- Commits to main trigger version evaluation
- Tags trigger release builds
- Pull requests create beta releases

2. Version numbers follow semantic versioning:
- MAJOR: Breaking changes
- MINOR: New features
- PATCH: Bug fixes

## Code Review Process

1. All changes require review
2. Address review comments promptly
3. Keep discussions focused and professional
4. Request re-review after making changes

## Community Guidelines

1. Be respectful and inclusive
2. Help others when possible
3. Follow the code of conduct
4. Keep discussions on-topic

## Getting Help

- Check existing documentation
- Search closed issues
- Ask in GitHub Discussions
- Join community chat

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
