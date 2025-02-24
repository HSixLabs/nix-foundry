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

## Package Issues

### Package Installation Fails

**Problem**: Package not found or installation fails.

**Solution**:
```bash
# Verify package name
nix-foundry config show packages

# Try installing with verbose output
nix-foundry install nodejs --verbose
```

### Package Conflicts

**Problem**: Package conflicts with existing installation.

**Solution**:
```bash
# Remove conflicting package
nix-foundry config set package remove nodejs

# Reinstall package
nix-foundry config set package add nodejs
nix-foundry config apply
```

## Configuration Issues

### Invalid Configuration

**Problem**: Configuration validation fails.

**Solution**:
```bash
# Reset to default configuration
nix-foundry config init --force

# Show current configuration
nix-foundry config show
```

### Shell Integration

**Problem**: Shell not properly configured.

**Solution**:
```bash
# Set shell in configuration
nix-foundry config set shell zsh

# Apply changes
nix-foundry config apply
```

## Platform-Specific Issues

### macOS

- Ensure Command Line Tools are installed
- Check SIP status
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

# List installed packages
nix-foundry config show packages

# Enable verbose logging
nix-foundry --verbose install nodejs
```

## Getting Help

- Run `nix-foundry --help` for command documentation
- Check command-specific help with `nix-foundry <command> --help`
- Review logs in `~/.config/nix-foundry/logs`
