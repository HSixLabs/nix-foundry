# Configuration Guide

Nix Foundry uses YAML configuration files to manage settings and scripts.

## Configuration File Location

Default location across all platforms:

`~/.config/nix-foundry/config.yaml`

This follows the XDG Base Directory Specification, providing a consistent configuration location regardless of the operating system.

## Configuration Format

```yaml
version: "1.0"
kind: config
metadata:
  name: nix-foundry-config
  description: Nix Foundry configuration
settings:
  autoUpdate: true
  updateInterval: 24h
  logLevel: info
nix:
  packages:
    core:
      - git
      - curl
    optional:
      - vim
      - tmux
  shell: /bin/zsh
  manager: nix-env
scripts:
  - name: setup-dev
    description: Setup development environment
    commands: |
      #!/usr/bin/env bash
      nix-env -iA nixpkgs.nodejs
      nix-env -iA nixpkgs.yarn
```

## Configuration Sections

### Metadata

Basic configuration information:

```yaml
metadata:
  name: string           # Configuration name
  description: string    # Configuration description
  version: string       # Configuration version
  created: timestamp    # Creation timestamp
  updated: timestamp    # Last update timestamp
```

### Settings

General application settings:

```yaml
settings:
  autoUpdate: boolean           # Enable automatic updates
  updateInterval: duration      # Update check interval
  logLevel: string             # Log level (debug|info|warn|error)
  colorOutput: boolean         # Enable colored output
  backupConfig: boolean        # Backup config before changes
  telemetry: boolean          # Enable anonymous usage data
  maxLogSize: string          # Maximum log file size
  maxLogFiles: number         # Number of log files to keep
```

### Nix Configuration

Nix-specific settings:

```yaml
nix:
  packages:
    core: string[]             # Essential packages
    optional: string[]         # Optional packages
  shell: string               # Default shell path
  manager: string             # Package manager (nix-env)
  multiUser: boolean          # Enable multi-user mode
  substituters: string[]      # Binary cache URLs
  trustedPublicKeys: string[] # Binary cache public keys
  maxJobs: number            # Maximum concurrent jobs
  cores: number              # Cores for building
  sandbox: boolean           # Enable build sandbox
```

### Shell Configuration

Shell-specific settings:

```yaml
shell:
  path: string               # Shell executable path
  rcFile: string            # RC file path
  completions: boolean      # Enable completions
  aliases: object           # Custom aliases
  environment: object       # Environment variables
  paths: string[]          # Additional PATH entries
```

### Scripts

Managed scripts:

```yaml
scripts:
  - name: string           # Script name
    description: string    # Script description
    commands: string      # Script content
    shell: string        # Script shell (optional)
    platform: string[]   # Supported platforms (optional)
    requires: string[]   # Required packages (optional)
```

## Environment Variables

Configuration can be overridden with environment variables:

```bash
export NIX_FOUNDRY_CONFIG=/path/to/config.yaml
export NIX_FOUNDRY_LOG_LEVEL=debug
export NIX_FOUNDRY_NO_COLOR=1
```

## Command Line Configuration

Quick configuration changes:

```bash
# Set a value
nix-foundry config set settings.autoUpdate true

# Get a value
nix-foundry config get settings.logLevel

# Reset to defaults
nix-foundry config reset

# Import/Export
nix-foundry config export backup.yaml
nix-foundry config import backup.yaml
```

## Platform-Specific Configuration

### macOS

```yaml
platform:
  darwin:
    sslCerts: boolean        # Manage SSL certificates
    brewIntegration: boolean # Homebrew integration
    architecture: string     # arm64 or x86_64
```

### Linux

```yaml
platform:
  linux:
    selinux: boolean        # SELinux support
    systemd: boolean        # systemd integration
    distribution: string    # Linux distribution
```

### WSL

```yaml
platform:
  wsl:
    windowsIntegration: boolean # Windows path integration
    mountPoint: string         # Windows mount point
    distro: string            # WSL distribution
```

## Configuration Examples

### Minimal Configuration

```yaml
version: "1.0"
kind: config
metadata:
  name: minimal-config
settings:
  autoUpdate: true
  logLevel: info
nix:
  packages:
    core: []
  shell: /bin/bash
```

### Development Environment

```yaml
version: "1.0"
kind: config
metadata:
  name: dev-environment
settings:
  autoUpdate: true
  logLevel: debug
nix:
  packages:
    core:
      - git
      - nodejs
      - python3
    optional:
      - docker
      - kubernetes-cli
  shell: /bin/zsh
scripts:
  - name: setup-dev
    description: Setup development environment
    commands: |
      #!/usr/bin/env bash
      nix-env -iA nixpkgs.vscode
      nix-env -iA nixpkgs.docker-compose
```

### Production Environment

```yaml
version: "1.0"
kind: config
metadata:
  name: prod-environment
settings:
  autoUpdate: false
  logLevel: warn
nix:
  packages:
    core:
      - monitoring-tools
      - security-tools
  multiUser: true
  sandbox: true
  maxJobs: 4
