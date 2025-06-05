# Packages

Nix Foundry provides access to the entire Nixpkgs collection through a simple YAML-based configuration system.

## Configuration

Define your packages in your configuration file:

```yaml
nix:
  packages:
    core:
      - git # Required packages
      - curl
    optional: # Additional packages
      - nodejs_20
      - any-nixpkgs-package
```

## Installation

During installation, Nix Foundry provides a wizard to help you select common development tools:

### Programming Languages

- Python (python311 with pip)
- Node.js (nodejs_20 with npm)
- Go (go_1_22 with gopls)
- Java (openjdk17 with maven)
- C/C++ (gcc with make)

### Development Tools

- Git (git)
- Docker (docker)
- Kubernetes CLI (kubectl)
- Terraform (terraform)
- GitHub CLI (gh)

### Editors

- VS Code (vscode)
- IntelliJ IDEA Community (jetbrains.idea-community)
- Neovim
- GNU Emacs
- Sublime Text

## Package Management

After installation, manage packages through the configuration system:

```bash
# Initialize configuration
nix-foundry config init

# Add packages to configuration
nix-foundry config set package add nodejs

# Apply configuration to install packages
nix-foundry config apply
```

## Platform Support

Packages are automatically handled based on your platform:

- Linux: Native package names
- macOS: Intel and ARM packages automatically selected
- WSL2: Linux packages with Windows integration

Remember: While we provide common development tools in the installation wizard, you can use any package from the Nixpkgs repository with Nix Foundry.
