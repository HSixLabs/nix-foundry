# Installation Guide

This guide covers installing Nix Foundry on different platforms and setting up Nix package manager.

## Automatic Installation

### macOS and Linux

```bash
curl -L https://github.com/shawnkhoffman/nix-foundry/releases/latest/download/nix-foundry-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o /usr/local/bin/nix-foundry
chmod +x /usr/local/bin/nix-foundry
```

### Windows (WSL)

```bash
curl -L https://github.com/shawnkhoffman/nix-foundry/releases/latest/download/nix-foundry-linux-amd64 -o /usr/local/bin/nix-foundry
chmod +x /usr/local/bin/nix-foundry
```

## Manual Installation

1. Download the appropriate binary for your platform from the [releases page](https://github.com/shawnkhoffman/nix-foundry/releases)
2. Move the binary to a location in your PATH
3. Make it executable with `chmod +x /path/to/nix-foundry`

## Installing Nix Package Manager

### Single-User Installation

```bash
nix-foundry install
```

### Multi-User Installation (macOS and Linux)

```bash
sudo nix-foundry install --multi-user
```

### WSL Installation

```bash
nix-foundry install
```

Note: WSL only supports single-user installation mode.

## Verifying Installation

```bash
nix-foundry --version
nix-env --version
```

## Shell Configuration

Nix Foundry automatically configures your shell environment. Supported shells:

- Bash
- Zsh
- Fish

The configuration is added to:

- `~/.bashrc` (Linux/WSL)
- `~/.bash_profile` (macOS)
- `~/.zshrc` (macOS default)
- `~/.config/fish/config.fish`

## Uninstallation

To remove Nix:

```bash
nix-foundry uninstall
```

For multi-user installations:

```bash
sudo nix-foundry uninstall
```

To remove Nix Foundry itself:

```bash
rm $(which nix-foundry)
```

## Troubleshooting

### Common Issues

1. Permission Denied
```bash
sudo chown -R $(whoami) /nix
```

2. Shell Not Configured
```bash
nix-foundry config setup
```

3. WSL Path Issues
```bash
nix-foundry config repair
```

### Logs

Logs are stored in:
- `~/.local/share/nix-foundry/logs` (single-user)
- `/var/log/nix-foundry` (multi-user)

## Next Steps

- Read the [Getting Started](getting-started.md) guide
- Configure your [shell environment](configuration.md#shell-configuration)
- Learn about [platform-specific features](platform-support.md)
