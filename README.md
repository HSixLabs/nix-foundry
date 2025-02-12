# nix-foundry

A robust foundation for cross-platform development environments using Nix. Build consistent, reproducible development setups across Darwin, Linux, and Windows with enterprise-grade tooling and automation.

## Why nix-foundry?

- 🏗️ **Production-Ready Foundation**: Pre-built infrastructure with cross-platform support
- 🚀 **Zero-Config Setup**: Automatic platform & user detection with smart defaults
- 🔄 **Enterprise Ready**: Multi-user support, quality gates, and CI/CD pipelines
- 🛠️ **Flexible Architecture**: Modular design for easy customization
- 📦 **Comprehensive Tooling**: Development, testing, and deployment tools included
- 🔧 **Battle-Tested**: Used in production across various team sizes

## Technical Foundation

1. **Core Architecture**:
   - Nix Flakes for reproducible builds
   - Home Manager for user environments
   - Platform-specific optimizations (Darwin/Linux/Windows)

2. **Key Features**:
   - Dynamic user detection and configuration
   - Cross-platform shell setup (ZSH/PowerShell)
   - Integrated development tools
   - Automated testing and releases

## Ready-Made Infrastructure

nix-foundry provides a complete development infrastructure:

1. **Cross-Platform Setup**:
   - Automatic platform detection and configuration
   - Pre-configured for Darwin, Linux, and Windows
   - WSL2 support built-in

2. **Development Environment**:
   - Shell configurations (ZSH/PowerShell)
   - Git setup with conventional commits
   - VSCode with recommended extensions
   - Direnv for project-specific environments

3. **Quality Tools**:
   - Pre-commit hooks for code quality
   - Automated testing workflows
   - Semantic versioning
   - Changelog generation

4. **Enterprise Features**:
   - Multi-user support with dynamic detection
   - Homebrew integration for macOS
   - Modular architecture for team customization
   - Local override support for individual preferences

Skip the weeks of setup and start with a production-ready foundation.

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

## Quick Start

1. **Install Nix**:

   ```bash
   sh <(curl -L https://nixos.org/nix/install) --daemon
   ```

2. **Set Required Variables**:

   ```bash
   export GITHUB_TOKEN="your-token"    # Required for installation
   export HOST="$(hostname)"           # Optional: custom hostname
   ```

3. **Bootstrap Environment**:

   ```bash
   curl -H "Authorization: token ${GITHUB_TOKEN}" \
        -L https://raw.githubusercontent.com/shawnkhoffman/nix-foundry/main/install.sh | \
        bash -s -- install
   ```

## Platform-Specific Setup

### macOS

- Homebrew packages install automatically
- System preferences configured
- iTerm2 settings applied on launch

### Linux

- System configuration applied automatically
- Reboot may be required for kernel changes
- WSL2 support included

### Windows (experimental)

- Requires WSL2 enabled
- PowerShell configuration included
- Some manual setup may be needed

## System Architecture

```shell
.
├── flake.nix         # Core configuration
├── home/             # User environment
│   ├── default.nix   # Base configuration
│   ├── git.nix      # Git settings
│   └── vscode.nix   # Editor config
├── modules/          # System modules
│   ├── darwin/      # macOS specific
│   ├── nixos/       # Linux specific
│   └── shared/      # Cross-platform
└── lib/             # Helper functions
```

Each component is designed for easy customization while maintaining stability.

## For Teams

nix-foundry provides:

1. **Standardized Environments**: Ensure all developers work with identical setups
2. **Quality Gates**: Pre-configured commit hooks, linting, and formatting
3. **CI/CD Pipeline**: Ready-to-use GitHub Actions workflows
4. **Cross-Platform Support**: Works seamlessly across different operating systems

## For Individuals

1. **Quick Setup**: Bootstrap a professional development environment in minutes
2. **Best Practices**: Industry-standard tools and workflows preconfigured
3. **Future-Proof**: Easy to extend and customize as needs grow
4. **Multiple Systems**: Use the same setup across all your machines

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

## Framework Design

nix-foundry is designed as a flexible framework that you can adapt to your needs:

### Core Components

1. **Platform Detection**: Automatic system detection and configuration
2. **User Management**: Dynamic user detection and setup
3. **Module System**: Pluggable architecture for easy customization

### Customization Points

1. **User Preferences**:
   - Shell configuration (ZSH/PowerShell)
   - Git settings
   - VSCode preferences
   - Local overrides via `~/.zshrc.local`

2. **System Modules**:
   - Add/remove packages in `modules/`
   - Modify platform-specific settings
   - Create custom modules

3. **Development Tools**:
   - Change linting rules
   - Modify CI/CD workflows
   - Adjust commit message standards

The framework provides sensible defaults but is designed to be forked and customized to match your team's needs.

## Contributing

We welcome contributions! Please follow these guidelines:

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```none
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

Types:

- `feat`: New feature (minor version)
- `fix`: Bug fix (patch version)
- `docs`: Documentation only
- `style`: Code style changes
- `refactor`: Code changes (no features/fixes)
- `perf`: Performance improvements
- `test`: Adding/fixing tests
- `chore`: Maintenance tasks

### Development Workflow

1. Clone the repository
2. Set up pre-commit hooks:

```bash
# Install pre-commit
nix-shell -p pre-commit nixpkgs-fmt

# Install the hooks
pre-commit install -t pre-commit -t commit-msg
```
