# Nix Foundry

Take your development environment anywhere with this simple YAML-based configuration manager. Nix Foundry manages extensible configurations at user, team, and project levels, ensuring consistent and reproducible development environments across platforms.

[![Go Report Card](https://goreportcard.com/badge/github.com/shawnkhoffman/nix-foundry)](https://goreportcard.com/report/github.com/shawnkhoffman/nix-foundry)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Overview

Nix Foundry simplifies development environment management:

- 🚀 Simple YAML configurations that work everywhere
- 📦 Extensible environment definitions in pure YAML
- 🔧 User, team, and project-specific settings
- 🌟 Consistent development environments
- 🛠️ Cross-platform compatibility

## Quick Start

```bash
# Install Nix Foundry
nix-foundry install

# Install packages
nix-foundry config set package add nodejs

# View configuration
cat ~/.config/nix-foundry/config.yaml # Woah, my config is portable!

# Apply configuration
nix-foundry apply

# Then, start a new terminal and see your environment!
```

## Features

- **Simple Configuration**: Pure YAML-based environment definitions
- **Portable Environments**: Take your development setup anywhere
- **Multi-Level Configs**: User, team, and project configurations
- **Extensible System**: Build on existing configurations
- **Cross-Platform**: Linux, macOS, and Windows (WSL2)

## Documentation

- [Quick Start Guide](docs/QUICKSTART.md)
- [Configuration Guide](docs/CONFIGURATION.md)
- [Package Management](docs/PACKAGES.md)
- [Platform Support](docs/PLATFORMS.md)
- [Command Reference](docs/COMMANDS.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [Contributing](docs/CONTRIBUTING.md)

## Platform Support

- **Linux**: Most major distributions
- **macOS**: Intel and Apple Silicon
- **Windows**: WSL2

## Example Configuration

```yaml
version: "v1"
kind: "NixConfig"
type: "project"
metadata:
  name: "web-dev"
  description: "Web development environment"
settings:
  shell: "zsh"
  autoUpdate: true
nix:
  packages:
    core:
      - git
      - nodejs
    optional:
      - postgresql
      - redis
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
