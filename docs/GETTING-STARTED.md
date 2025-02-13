# Getting Started with nix-foundry

## Installation

```bash
# Install nix-foundry
curl -L https://get.nix-foundry.dev | sh

# Initialize your environment
nix-foundry init
```

## Personal Setup (5 minutes)

1. Create your config:
```bash
nix-foundry config init
```

2. Add essential tools:
```yaml
# ~/.config/nix-foundry/config.yaml
shell:
  type: zsh          # or bash/fish
  plugins:
    - zsh-autosuggestions

packages:
  - git
  - neovim
  - ripgrep
```

3. Apply changes:
```bash
nix-foundry apply
```

## Team Project Setup (2 minutes)

In your project directory:
```bash
# Initialize project environment
nix-foundry project init

# Import team settings
nix-foundry project import
```

## Essential Commands

### Environment Management
```bash
# Switch environments
nix-foundry switch personal  # Use personal setup
nix-foundry switch project   # Use team setup

# Update and maintain
nix-foundry update          # Update everything
nix-foundry doctor          # Check for issues
nix-foundry rollback        # Revert to last working state
```

### Configuration
```bash
# Initialize configs
nix-foundry config init     # Create personal config
nix-foundry project init    # Create project config

# Manage configs
nix-foundry config validate # Check configuration
nix-foundry project import  # Import team settings
nix-foundry apply          # Apply changes
```

### Troubleshooting
```bash
nix-foundry doctor         # Diagnose issues
nix-foundry logs          # View logs
nix-foundry clean         # Clean environment
nix-foundry status        # Check current state
```

## What's Next?

1. **Add More Tools**
   - [Configuration Guide](CONFIG.md)
   - [Package List](https://nixos.org/nixpkgs)

2. **Team Integration**
   - [Team Setup Guide](TEAM.md)
   - [Best Practices](docs/TEAM.md#best-practices)

3. **Get Help**
   - Run `nix-foundry help`
   - Check [FAQ](FAQ.md)
   - Join [Discord](https://discord.gg/nix-foundry)
