# nix-configs

A comprehensive Nix configuration system supporting Darwin (macOS), NixOS, and Windows (experimental) environments.

## Features

- 🚀 Multi-platform support (Darwin, NixOS, WSL, Windows)
- 🏠 Home Manager integration
- 🍺 Homebrew integration for macOS
- 🔧 Automated development environment setup
- 🐚 ZSH configuration with Powerlevel10k
- 📦 Consistent package management across systems

## Supported Systems

- aarch64-darwin (Apple Silicon macOS)
- x86_64-darwin (Intel macOS)
- aarch64-linux
- x86_64-linux
- x86_64-windows (experimental)

## Prerequisites

- Nix package manager installed
- Git installed
- GitHub personal access token with repo scope (set as `GITHUB_TOKEN` environment variable)
- For macOS: Xcode Command Line Tools

## Installation & Updates

This configuration uses an install script that handles all installation and maintenance operations.

Using curl:

```bash
# Choose one of: install, update, or reinstall
curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- <operation>
```

Or using wget:

```bash
# Choose one of: install, update, or reinstall
wget -qO- --header="Authorization: token ${GITHUB_TOKEN}" https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- <operation>
```

For example, using either curl:

```bash
# Fresh installation
curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- install

# Update existing installation
curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- update

# Clean reinstall
curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- reinstall
```

Or wget:

```bash
# Fresh installation
wget -qO- --header="Authorization: token ${GITHUB_TOKEN}" https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- install

# Update existing installation
wget -qO- --header="Authorization: token ${GITHUB_TOKEN}" https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- update

# Clean reinstall
wget -qO- --header="Authorization: token ${GITHUB_TOKEN}" https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- reinstall
```

The install script handles:

- Fresh installation setup
- Updates to existing installations (preserving local changes)
- Clean reinstallation when needed
- Platform-specific setup (macOS, NixOS, Windows)
- Configuration file management
- Home Manager installation
- Homebrew setup (macOS only)

Operations:

- `install`: Performs a fresh installation
- `update`: Updates existing installation while preserving local changes
- `reinstall`: Removes existing configuration and performs a clean reinstall

> **Note**: The `reinstall` operation will remove the current configuration. Make sure to backup any local changes before using this option.

## Post-install Configuration

### Shell Configuration

Configure Powerlevel10k:

```bash
POWERLEVEL9K_CONFIG_FILE="$HOME/.config/zsh/conf.d/p10k.zsh" p10k configure
```

### Platform-Specific Setup

#### macOS

- Automatically configures SSL certificates
- Sets up Homebrew and manages packages
- Configures system defaults and fonts
- Integrates with iTerm2

#### NixOS

- Full system configuration
- GNOME desktop environment
- Hardware and network configuration
- Development tools and utilities

#### Windows (Experimental)

- WSL integration
- PowerShell configuration
- Windows-specific path handling

## Directory Structure

```bash
.
├── flake.nix           # Main flake configuration
├── home/               # Home-manager configurations
├── lib/                # Helper functions and utilities
├── modules/            # System modules
│   ├── darwin/         # macOS-specific configurations
│   ├── nixos/         # NixOS-specific configurations
│   ├── shared/        # Cross-platform configurations
│   └── windows/       # Windows-specific configurations
└── users.nix          # User configuration
```

## Customization

### Adding New Users

Edit `users.nix` to add new users or modify existing ones. The system will automatically detect the current user and configure accordingly.

### Adding New Packages

- System-wide packages: Add to respective module files in `modules/`
- User-specific packages: Add to configurations in `home/`
- macOS-specific packages: Add to `modules/darwin/homebrew.nix`

## Contributing

We use conventional commits for automated semantic versioning. Please see our
[Contributing Guidelines](CONTRIBUTING.md) for details on:

- Commit message format
- Development workflow
- Testing requirements
- Release process

## License

This project is licensed under the MIT License - see the LICENSE file for details.
