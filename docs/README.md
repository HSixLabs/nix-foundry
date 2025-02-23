# Nix Foundry Documentation

Welcome to the Nix Foundry documentation. This tool helps you manage Nix installations across different platforms with a consistent interface.

## Table of Contents

- [Installation](installation.md)
- [Getting Started](getting-started.md)
- [Configuration](configuration.md)
- [Commands](commands.md)
- [Platform Support](platform-support.md)
- [Contributing](contributing.md)
- [Troubleshooting](troubleshooting.md)

## Quick Start

```bash
# Install nix-foundry
curl -L https://github.com/shawnkhoffman/nix-foundry/releases/latest/download/nix-foundry-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o /usr/local/bin/nix-foundry
chmod +x /usr/local/bin/nix-foundry

# Initialize nix-foundry
nix-foundry init

# Install Nix package manager
nix-foundry install

# View available commands
nix-foundry help
```

## Features

- Cross-platform support (macOS, Linux, WSL)
- Multi-user and single-user installation modes
- Automated shell configuration
- Platform-specific optimizations
- Comprehensive test coverage
- Clean uninstallation process

## System Requirements

- macOS 10.15 or later
- Linux (any modern distribution)
- Windows Subsystem for Linux (WSL2)
- 1GB free disk space
- Internet connection for installation

## Support

For bug reports and feature requests, please use the [GitHub issue tracker](https://github.com/shawnkhoffman/nix-foundry/issues).

## License

MIT License - see [LICENSE](../LICENSE) for details.
