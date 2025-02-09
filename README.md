# nix-configs

My Nix configuration system supporting Darwin (macOS), NixOS, and Windows (experimental) environments.

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
- aarch64-linux (Apple Silicon Linux)
- x86_64-linux (Intel Linux)
- x86_64-windows (experimental)

## Prerequisites

- Nix package manager
- Git
- GitHub personal access token (set as `GITHUB_TOKEN` environment variable)
- For macOS: Xcode Command Line Tools

## Installation

First, install Nix if you haven't already:

```bash
# macOS/Linux
sh <(curl -L https://nixos.org/nix/install) --daemon

# Windows (requires WSL2)
sh <(curl -L https://nixos.org/nix/install) --no-daemon
```

Then install the configuration:

```bash
# Fresh install
curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash -s -- install
```

Other operations:

- `update`: Updates existing installation
- `reinstall`: Performs clean reinstall (warning: removes current config)

## Post-install Setup

1. Configure Powerlevel10k:

    ```bash
    POWERLEVEL9K_CONFIG_FILE="$HOME/.config/zsh/conf.d/p10k.zsh" p10k configure
    ```

2. For macOS:

    - Homebrew packages will be installed automatically
    - System defaults will be configured
    - iTerm2 settings will be applied on next launch

3. For NixOS:

    - System configuration will be applied automatically
    - Reboot may be required for some changes

4. For Windows:

    - WSL2 must be enabled
    - Some features may require manual configuration

## Directory Structure

```shell
.
├── .github/           # GitHub Actions and configs
├── flake.nix           # Main flake configuration
├── home/               # Home-manager configurations
├── lib/                # Helper functions and utilities
├── modules/            # System modules
│   ├── darwin/         # macOS-specific configurations
│   ├── nixos/         # NixOS-specific configurations
│   ├── shared/        # Cross-platform configurations
│   └── windows/       # Windows-specific configurations
├── users.nix           # User configuration
├── install.sh          # Installation script
└── flake.lock          # Flake dependencies lock file
```

## Customization

Add packages to:

- `modules/`: System-wide packages
- `home/`: User-specific packages
- `modules/darwin/homebrew.nix`: macOS-specific packages

Local customizations:

- ZSH: Create `~/.zshrc.local` for machine-specific settings
- Git: Edit `home/git.nix` for Git configuration
- VSCode: Modify `home/vscode.nix` for editor settings

Note: The configuration automatically detects your username and sets up accordingly - no need to modify `users.nix` unless you want to change the default behavior.

## Contributing

While this is my personal configuration, I welcome contributions! You can:

- 🔄 Fork it as a starting point for your own config
- 🔍 Use it as a reference or template for your own config
- 🐛 Report issues if you find them
- 💡 Suggest improvements
- 🤝 Submit PRs for bugs or enhancements

For consistency, please use conventional commits when contributing:

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `refactor`: Code changes that neither fix bugs nor add features

See [Contributing Guidelines](CONTRIBUTING.md) for more details.

## License

MIT License - see the [LICENSE](LICENSE) file for details.
