# Team Customization Guide

## Overview

This guide explains how to customize nix-foundry for advanced team needs beyond the basic team configuration.

## Extension Points

### 1. Custom Package Sets

Define team-specific package collections in `~/.config/nix-foundry/team-packages.yaml`:

```yaml
development:
  backend:
    - terraform
    - kubectl
    - aws-cli
    - gopls
    - delve
  frontend:
    - nodejs
    - yarn
    - eslint
    - prettier
  database:
    - postgresql
    - pgcli
    - redis-cli
```

### 2. Advanced Team Configuration

Create specialized team profiles:

```yaml
# team-profile.yaml
name: "backend-team"
extends: "base-go-team"  # Inherit from base profile

overrides:
  packages:
    required:
      - terraform
      - aws-cli
    optional:
      - localstack
      - k9s

  tools:
    formatting:
      go: gofmt
      terraform: terraform fmt
    linting:
      go: golangci-lint
      terraform: tflint

  ci:
    checks:
      - security-scan
      - license-check
      - dependency-audit
```

## Templates

nix-foundry provides customizable templates in `~/.config/nix-foundry/templates/`:

```shell
templates/
├── flake/              # Nix flake templates
│   ├── go.nix         # Go development
│   ├── node.nix       # Node.js development
│   └── python.nix     # Python development
├── configs/           # Configuration templates
│   ├── backend/
│   └── frontend/
└── ci/               # CI/CD templates
    ├── github/
    └── gitlab/
```

## Health Checks

Add custom health checks in `~/.config/nix-foundry/health-checks.yaml`:

```yaml
checks:
  database:
    - name: "PostgreSQL Connection"
      command: "pg_isready"
      interval: "5m"

  services:
    - name: "Redis Status"
      command: "redis-cli ping"
      interval: "1m"

  tools:
    - name: "Go Tools"
      verify:
        - "go version"
        - "gopls version"
        - "golangci-lint version"
```

## Platform-Specific Customization

### macOS

```yaml
platform:
  darwin:
    homebrew:
      taps:
        - hashicorp/tap
      casks:
        - docker
        - visual-studio-code
    defaults:
      dock:
        autohide: true
```

### Linux

```yaml
platform:
  linux:
    systemPackages:
      - docker-ce
      - build-essential
    sysctl:
      vm.max_map_count: 262144
      fs.inotify.max_user_watches: 524288
```

## Configuration Structure

nix-foundry uses the following directory structure:

```shell
~/.config/nix-foundry/
├── flake.nix          # Generated Nix flake
├── home.nix           # Home-manager configuration
├── packages.json      # Package list
├── backups/           # Backup archives
└── templates/         # Custom flake templates
```

## Development Workflow

```bash
# Initialize environment
nix-foundry init --auto

# Add development packages
nix-foundry packages add gopls delve golangci-lint

# Apply changes
nix-foundry apply

# Update environment
nix-foundry update
```

## Backup and Restore

Manage your team's configurations:

```bash
# Create backup
nix-foundry backup create

# List backups
nix-foundry backup list

# Restore configuration
nix-foundry backup restore <backup-file>

# Delete backup
nix-foundry backup delete <backup-file>
```

## Health Checks

The `doctor` command verifies your setup:

```bash
nix-foundry doctor
```

This checks:
- Nix installation
- home-manager presence
- Flakes configuration
- Directory structure
- Configuration integrity

## Troubleshooting

Common issues and solutions:

1. **Configuration Issues**
   - Run `nix-foundry doctor` to check setup
   - Verify directory permissions
   - Check configuration files exist

2. **Package Issues**
   - Verify package names in packages.json
   - Check Nix flake lock file
   - Run `nix-foundry update` to refresh

3. **Backup/Restore Issues**
   - Ensure backup directory exists
   - Check file permissions
   - Verify sufficient disk space

## Prerequisites

### Nix Configuration

1. **Enable Flakes**
Add to `~/.config/nix/nix.conf`:

    ```shell
    experimental-features = nix-command flakes
    ```

2. **Default Flake Structure**
nix-foundry generates a `flake.nix` with these inputs:

    ```nix
    {
    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
        home-manager = {
        url = "github:nix-community/home-manager";
        inputs.nixpkgs.follows = "nixpkgs";
        };
        # Darwin-specific inputs
        nix-darwin = {
        url = "github:LnL7/nix-darwin";
        inputs.nixpkgs.follows = "nixpkgs";
        };
    };
    }
    ```

### Platform-Specific Setup

#### Linux

```bash
# Install Nix
curl -L https://nixos.org/nix/install | sh

# Enable flakes
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf
```

#### macOS

```bash
# Install Nix
curl -L https://nixos.org/nix/install | sh

# Install nix-darwin
nix-build https://github.com/LnL7/nix-darwin/archive/master.tar.gz -A installer
./result/bin/darwin-installer

# Enable flakes
echo "experimental-features = nix-command flakes" | sudo tee -a /etc/nix/nix.conf
```
