# Configuration

Nix Foundry uses a hierarchical YAML-based configuration system.

## Schema

```yaml
version: string # v1
kind: string # NixConfig
type: string # user|team|project
metadata:
  name: string # Configuration name
  description: string # Configuration description
  created: timestamp # Creation timestamp
  updated: timestamp # Last update timestamp
  priority?: number # Higher priority configs override lower ones
base?: string # Name of the config to extend from
settings:
  shell: string # bash|zsh|fish
  logLevel: string # info|debug|warn|error
  autoUpdate: boolean
  updateInterval: duration # e.g., 24h
nix:
  manager: string # nix-env
  packages:
    core?: [string] # Required for team/project configs
    optional?: [string]
  scripts?:
    - name: string
      description?: string
      commands: string # Multiline string with | style
```

## File Locations

- User config: `~/.config/nix-foundry/config.yaml`
- Team configs: `~/.config/nix-foundry/teams/<name>.yaml`
- Project config: `./.nix-foundry/config.yaml`

## Example Configuration

```yaml
version: 'v1'
kind: 'NixConfig'
type: 'user'
metadata:
  name: 'default'
  description: 'Default user configuration'
settings:
  shell: 'zsh'
  logLevel: 'info'
  autoUpdate: true
  updateInterval: '24h'
nix:
  manager: 'nix-env'
  packages:
    core:
      - git
      - curl
    optional:
      - nodejs
  scripts:
    - name: 'setup-dev'
      description: 'Set up development environment'
      commands: |
        nix-env -iA nixpkgs.nodejs
        npm install -g yarn
```

## Management Commands

### Initialize

```bash
# User configuration
nix-foundry config init

# Team configuration
nix-foundry config init --type team --name myteam

# Project configuration
nix-foundry config init --type project
```

### View

```bash
# Show current configuration
nix-foundry config show

# List all configurations
nix-foundry config list
```

### Modify

```bash
# Set shell preference
nix-foundry config set shell zsh

# Add a package
nix-foundry config set package add nodejs

# Add a script
nix-foundry config set script add setup-dev
```

## Configuration Hierarchy

1. Project configuration (highest priority)
2. Team configuration
3. User configuration (lowest priority)

Settings are merged with higher priority configurations overriding lower ones.
