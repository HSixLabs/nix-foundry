# Team Environment Guide

## Quick Setup

```bash
# In your project root
nix-foundry project init
nix-foundry project import
```

For configuration examples and options, see:
- [Configuration Guide](CONFIG.md)
- [Configuration Reference](CONFIG-REFERENCE.md)

## Common Tasks

### Switch Environments
```bash
# Use team environment
nix-foundry switch project

# Temporarily use personal setup
nix-foundry switch personal
```

### Update Team Settings
```bash
# Pull latest team config
nix-foundry project update

# Check for issues
nix-foundry doctor
```

## Best Practices

See our comprehensive [Best Practices Guide](BEST-PRACTICES.md#team-setup).

## Troubleshooting

For detailed troubleshooting steps, see:
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)
