# nix-foundry

A robust foundation for cross-platform development environments using Nix. Build consistent, reproducible development setups across Darwin, Linux, and Windows with enterprise-grade tooling and automation.

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

- 🏗️ **Production-Ready**: Cross-platform support with automatic detection
- 🚀 **Zero-Config**: Smart defaults with dynamic user detection
- 🔄 **Enterprise-Grade**: Multi-user support, quality gates, CI/CD
- 🛠️ **Development Tools**: Git, VSCode, Shell configurations included
- 📦 **Quality Tools**: Pre-commit hooks, testing, semantic versioning
- 🔧 **Customizable**: Modular design for team adaptation

## Supported Systems

- macOS (Apple Silicon & Intel)
- Linux (x86_64 & ARM)
- Windows (experimental, via WSL2)

## Customization

- Add packages to `modules/` (system) or `home/` (user)
- Create `~/.zshrc.local` for machine-specific settings
- Modify Git/VSCode settings in respective `.nix` files

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Platform Setup](docs/PLATFORMS.md)
- [Development](docs/DEVELOPMENT.md)
- [Contributing](CONTRIBUTING.md)
