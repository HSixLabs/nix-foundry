# Quick Start Guide

## Installation

1. Install Nix Foundry:
   ```bash
   nix-foundry install
   ```

2. Verify installation:
   ```bash
   nix-foundry --version
   ```

## Basic Usage

### Initialize Configuration

```bash
nix-foundry config init
```

This creates a default configuration in `~/.config/nix-foundry/config.yaml`.

### Install Packages

1. Install a single package:
   ```bash
   nix-foundry config set package add nodejs
   ```

2. Install from configuration:
   ```bash
   nix-foundry config apply
   ```

## Common Tasks

### Managing Packages

- List available packages:
  ```bash
  nix-foundry config show packages
  ```

- Add a package to configuration:
  ```bash
  nix-foundry config set package add nodejs
  ```

### Managing Shell

- Set default shell:
  ```bash
  nix-foundry config set shell zsh
  ```

- Apply shell changes:
  ```bash
  nix-foundry config apply
  ```

## Next Steps

- Read the [full documentation](./README.md)
- Configure your [development environment](./CONFIGURATION.md)
- Explore available [packages](./PACKAGES.md)
- Check [platform support](./PLATFORMS.md)

## Getting Help

- Run `nix-foundry --help` for command documentation
- Visit our [troubleshooting guide](./TROUBLESHOOTING.md)
