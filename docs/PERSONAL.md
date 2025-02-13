# Personal Environment Guide

For common commands and setup, see [Getting Started](GETTING-STARTED.md#essential-commands).

## Quick Setup

```bash
# Initialize environment
nix-foundry init
```

For configuration examples and options, see:
- [Configuration Guide](CONFIG.md)
- [Configuration Reference](CONFIG-REFERENCE.md)

## Common Tasks

### Profile Management
```bash
# Create new profile
nix-foundry profile create work

# Switch profiles
nix-foundry switch personal
nix-foundry switch work
```

### Backup/Restore
```bash
# Create backup
nix-foundry backup create

# Restore from backup
nix-foundry backup restore latest
```

## Best Practices

See our comprehensive [Best Practices Guide](BEST-PRACTICES.md#personal-setup).

## Troubleshooting

For detailed troubleshooting steps, see:
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)
