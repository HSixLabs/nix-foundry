# Troubleshooting Guide

## Quick Fixes

### 1. Environment Issues
```bash
# Fix most common issues
nix-foundry doctor --fix

# Reset to last working state
nix-foundry rollback
```

### 2. Configuration Problems
```bash
# Validate configuration
nix-foundry config validate

# Show conflicts
nix-foundry config check-conflicts
```

## Common Issues

### Installation Fails
1. Check prerequisites:
```bash
# Verify Nix installation
nix --version

# Check home-manager
home-manager --version
```

2. Try clean install:
```bash
nix-foundry uninstall --clean
curl -L https://get.nix-foundry.dev | sh
```

### Platform-Specific Issues

#### Linux
- SELinux conflicts: `nix-foundry setup linux --selinux`
- Permission issues: Check user permissions
- System packages: Run `nix-foundry doctor --system`

#### macOS
- Rosetta 2: `nix-foundry setup macos --rosetta`
- SIP limitations: `nix-foundry setup macos --sip`
- Homebrew conflicts: Run `nix-foundry doctor --brew`

#### Windows (WSL2)
- WSL2 setup: `nix-foundry setup wsl`
- Path issues: `nix-foundry doctor --wsl`
- Performance: Enable WSL2 optimizations

### Environment Problems
1. Check status:
```bash
nix-foundry status
```

2. Common fixes:
```bash
# Clean environment
nix-foundry clean

# Force switch
nix-foundry switch project --force
```

## Getting Help

### 1. Collect Information
```bash
# Create diagnostic report
nix-foundry doctor > diagnostic.log
nix-foundry logs > logs.txt
```

### 2. Get Support
- Check [FAQ](FAQ.md)
- Join [Discord](https://discord.gg/nix-foundry)
- Open [Issue](https://github.com/shawnkhoffman/nix-foundry/issues)
