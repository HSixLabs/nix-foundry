# Contributing Guide

## Quick Start

```bash
# Setup development environment
git clone https://github.com/shawnkhoffman/nix-foundry.git
cd nix-foundry
nix-foundry dev init

# Run tests
nix-foundry test
```

## Development Workflow

### 1. Prepare
```bash
# Create feature branch
git checkout -b feature/your-feature

# Enable development tools
nix-foundry dev enable
```

### 2. Develop
```bash
# Run tests while developing
nix-foundry test --watch

# Check formatting
nix-foundry fmt check
```

### 3. Submit
```bash
# Create PR
nix-foundry pr create --type feature
```

## Guidelines

### Code Style
- Follow Go standards
- Document public APIs
- Write clear commit messages
- Include tests

### Testing
```bash
# Run all tests
nix-foundry test

# Test specific component
nix-foundry test --component config
```

### Documentation
- Update relevant docs
- Add examples for new features
- Keep README.md current

## Release Process

1. **Prepare Release**
   - Update CHANGELOG.md
   - Bump version numbers
   - Update documentation

2. **Submit**
   - Create release PR
   - Address review feedback
   - Wait for CI approval

Need help? See:
- [Development Guide](DEVELOPMENT.md)
- [FAQ](FAQ.md)
- Join [Discord](https://discord.gg/nix-foundry)
