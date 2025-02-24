# Troubleshooting

Common issues and solutions when using Nix Foundry.

## Installation Issues

### Multi-User Installation

**Problem**: Permission denied during multi-user installation.

**Solution**:

```bash
# Run with sudo for multi-user mode
sudo nix-foundry install --multi-user
```

### Single-User Installation

**Problem**: Configuration directory not accessible.

**Solution**:

```bash
# Check directory permissions
ls -la ~/.config/nix-foundry
# Fix permissions if needed
chmod 755 ~/.config/nix-foundry
```

## Configuration Issues

### Invalid Configuration

**Problem**: Configuration validation fails.

**Solution**:

```bash
# Reset to default configuration
nix-foundry config init

# Show current configuration
nix-foundry config show
```

### Configuration Not Found

**Problem**: Configuration file missing.

**Solution**:

```bash
# Initialize new configuration
nix-foundry config init

# For team configuration
nix-foundry config init --type team --name myteam

# For project configuration
nix-foundry config init --type project
```

## Package Issues

### Package Installation Fails

**Problem**: Package installation through config apply fails.

**Solution**:

1. Verify configuration:

   ```bash
   nix-foundry config show
   ```

2. Try applying configuration:
   ```bash
   nix-foundry config apply
   ```

### Shell Integration

**Problem**: Shell not properly configured.

**Solution**:

1. Set shell in configuration:

   ```bash
   nix-foundry config set shell zsh
   ```

2. Apply changes:
   ```bash
   nix-foundry config apply
   ```

## Platform-Specific Issues

### macOS

- Ensure Command Line Tools are installed
- Multi-user mode is required
- Verify Rosetta 2 for ARM systems

### Linux

- Check systemd status for multi-user mode
- Verify package compatibility
- Check filesystem permissions

### WSL2

- Ensure WSL2 is properly configured
- Use Linux filesystem for better performance
- Check Windows integration status

## Common Commands

```bash
# Show version and system info
nix-foundry --version

# Show detailed configuration
nix-foundry config show

# List available configurations
nix-foundry config list

# Enable verbose logging
nix-foundry --verbose install
```

## Getting Help

- Run `nix-foundry --help` for command documentation
- Check command-specific help with `nix-foundry <command> --help`
- Review configuration in `~/.config/nix-foundry/config.yaml`
