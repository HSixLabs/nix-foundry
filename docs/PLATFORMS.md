# Platform Support

## Quick Reference

| Platform | Status | Notes |
|----------|---------|-------|
| Linux    | Full    | All features supported |
| macOS    | Full    | Intel & Apple Silicon |
| Windows  | Beta    | WSL2 recommended |

## Platform-Specific Setup

### Linux
```bash
# Install dependencies
nix-foundry setup linux

# Configure SELinux (if needed)
nix-foundry setup linux --selinux
```

### macOS
```bash
# Install with Rosetta 2 support
nix-foundry setup macos --rosetta

# Configure SIP exceptions
nix-foundry setup macos --sip
```

### Windows (WSL2)
```bash
# Setup WSL environment
nix-foundry setup wsl

# Enable Windows integration
nix-foundry setup wsl --windows-tools
```

## Common Issues

For platform-specific troubleshooting, see:
- [Troubleshooting Guide](TROUBLESHOOTING.md#platform-specific-issues)
- [FAQ](FAQ.md#platform-specific)

## Best Practices

### Cross-Platform Development
- Use platform-agnostic paths
- Test on all platforms
- Handle platform errors

### Performance
- Enable platform optimizations
- Use native binaries
- Cache aggressively

Need help? See:
- [Troubleshooting](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)
- [Security Guide](SECURITY.md)
