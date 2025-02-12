# nix-foundry

A framework for building consistent, reproducible development environments across platforms. nix-foundry helps teams standardize their development setups using Nix, with enterprise-grade tooling and automation.

## Quick Start

1. Install Nix:

   ```bash
   sh <(curl -L https://nixos.org/nix/install) --daemon
   ```

2. Bootstrap your environment:

   ```bash
   export GITHUB_TOKEN="your-token"
   curl -H "Authorization: token ${GITHUB_TOKEN}" \
        -L https://raw.githubusercontent.com/shawnkhoffman/nix-foundry/main/install.sh | \
        bash -s -- install
   ```

## Features

- 🏗️ **Cross-Platform**: Consistent environments across macOS, Linux, and Windows (WSL2)
- 🚀 **Zero-Config**: Smart defaults with automatic platform detection
- 🔄 **Enterprise-Ready**: Multi-user support, quality gates, CI/CD integration
- 🛠️ **Development Tools**: Pre-configured Git, VSCode, Shell environments
- 📦 **Quality Assurance**: Pre-commit hooks, testing, semantic versioning
- 🔧 **Extensible**: Modular design for team customization

## Supported Platforms

- macOS (Apple Silicon & Intel)
- Linux (x86_64 & ARM)
- Windows (via WSL2)

## Customization

- Add packages to `modules/` (system) or `home/` (user)
- Create `~/.zshrc.local` for machine-specific settings
- Modify Git/VSCode settings in respective `.nix` files

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Platform Setup](docs/PLATFORMS.md)
- [Development](docs/DEVELOPMENT.md)
- [Contributing](CONTRIBUTING.md)
