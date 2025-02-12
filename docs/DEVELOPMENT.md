# Development Guide

This guide covers the development workflow and testing process for nix-foundry.

## Development Environment

1. Install Nix package manager
2. Clone the repository
3. Enter the development environment and install hooks:

   ```bash
   nix-shell  # This loads the development environment
   pre-commit install
   ```

## Quality Gates

The project enforces several quality gates during development:

1. **Pre-commit Hooks**
   - Code formatting
   - Syntax checking
   - Commit message validation

2. **Build Validation**

   ```bash
   # Check flake builds
   nix flake check

   # Test home-manager configuration
   nix run home-manager/master -- switch \
     --flake .#default \
     --extra-experimental-features "nix-command flakes" \
     --impure
   ```

   See [install.sh:341-354](../install.sh) for implementation details.

## Testing

### Platform Testing

The CI pipeline automatically tests builds across supported platforms:

- Ubuntu (ARM64 & x86_64)
- macOS (Apple Silicon & Intel)
- Windows (experimental, via WSL2)

Tests run on pull requests and main branch pushes. See [test.yml:64-104](.github/workflows/test.yml) for implementation details.

### Local Testing

Test your changes locally before submitting:

1. **Syntax Check**

   ```bash
   nix-shell -p nixpkgs-fmt --run "nixpkgs-fmt *.nix"
   ```

2. **Build Test**

   ```bash
   nix build .#homeConfigurations.$USER-$(uname -m)-$(uname -s | tr '[:upper:]' '[:lower:]').activationPackage
   ```

## Release Process

Releases are automated using semantic versioning based on conventional commits:

1. Merge changes to `main`
2. CI evaluates commit messages
3. Version is bumped according to changes:
   - Breaking changes: major version
   - New features: minor version
   - Bug fixes: patch version
4. Changelog is generated
5. Release is created with notes

See [CONTRIBUTING.md:107-119](../CONTRIBUTING.md) for more details.

## Debugging

1. Enable debug logging:

   ```bash
   export NIX_DEBUG=1
   ```

2. Check build outputs:

   ```bash
   nix log /nix/store/<hash>
   ```

3. Test specific platforms:

   ```bash
   nix build .#darwinConfigurations.default.system  # macOS
   nix build .#homeConfigurations.default.activationPackage  # Linux/WSL
   ```

## Best Practices

1. Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification
2. Test changes across all supported platforms
3. Update documentation for significant changes
4. Add comments for complex configurations
5. Use descriptive commit messages
