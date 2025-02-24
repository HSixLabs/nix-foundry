# Commands

Nix Foundry provides a comprehensive CLI with various commands for managing your Nix environment.

## Command Documentation

Detailed documentation for each command is available in the [commands](./commands) directory.

## Core Commands

- `nix-foundry install` - Install Nix package manager
- `nix-foundry config` - Manage Nix Foundry configuration
- `nix-foundry uninstall` - Uninstall Nix Foundry

## Configuration Commands

- `nix-foundry config init` - Initialize a new configuration
- `nix-foundry config apply` - Apply the current configuration
- `nix-foundry config list` - List available configurations
- `nix-foundry config set` - Set configuration values
- `nix-foundry config show` - Show configuration details

## Common Options

All commands support:
- `--verbose, -v` - Enable verbose output
- `--help, -h` - Show help for any command

## Usage Examples

```bash
# Show command help
nix-foundry --help
nix-foundry <command> --help

# Initialize configuration
nix-foundry config init

# Install packages
nix-foundry install nodejs
```

For detailed documentation of each command, please see the [commands](./commands) directory.
