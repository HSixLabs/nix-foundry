# Platform Support

Nix Foundry provides comprehensive support for multiple platforms with platform-specific optimizations and features.

## macOS

### Supported Versions
- macOS 10.15 (Catalina) and later
- Both Intel (x86_64) and Apple Silicon (arm64) architectures

### Features
- Multi-user installation support
- Native zsh shell integration
- Homebrew compatibility
- System Integrity Protection (SIP) handling
- Automatic PATH configuration
- Apple Silicon native performance

### Installation Modes
- Single-user: Standard installation in user's home directory
- Multi-user: System-wide installation with daemon support

### Configuration Paths
- Nix Store: `/nix`
- Configuration: `/etc/nix`
- User Profile: `~/.nix-profile`
- Shell Config: `~/.zshrc` or `~/.bash_profile`

## Linux

### Supported Distributions
- Ubuntu 20.04 and later
- Debian 10 and later
- Fedora 34 and later
- CentOS/RHEL 8 and later
- Arch Linux
- NixOS (native support)

### Features
- Multi-user installation support
- SELinux compatibility
- systemd integration
- Multiple shell support (bash, zsh, fish)
- Distribution-specific optimizations

### Installation Modes
- Single-user: Standard installation in user's home directory
- Multi-user: System-wide installation with daemon support

### Configuration Paths
- Nix Store: `/nix`
- Configuration: `/etc/nix`
- User Profile: `~/.nix-profile`
- Shell Config: `~/.bashrc` or `~/.zshrc`

## Windows Subsystem for Linux (WSL)

### Requirements
- WSL2
- Any supported Linux distribution under WSL
- Windows 10 version 2004 or later

### Features
- Single-user installation only (by design)
- Windows/Linux path conversion
- Automatic Windows binary wrapping
- VSCode integration
- Windows filesystem access

### Limitations
- No multi-user support
- No native Windows support (WSL required)
- Limited systemd support (distribution dependent)

### Configuration Paths
- Nix Store: `/nix`
- Configuration: `/etc/nix`
- User Profile: `~/.nix-profile`
- Shell Config: `~/.bashrc`

## Common Features Across Platforms

### Shell Integration
- Automatic PATH configuration
- Shell completion setup
- Environment variable configuration
- Profile backup and restoration

### Package Management
- Binary cache support
- Garbage collection
- Profile management
- Rollback support

### Security
- Sandboxed builds
- Binary verification
- Secure downloads
- Permission management

## Platform-Specific Commands

### macOS
```bash
# Enable multi-user support
sudo nix-foundry install --multi-user

# Configure for Apple Silicon
nix-foundry config set architecture arm64
```

### Linux
```bash
# SELinux configuration
nix-foundry config set selinux-support true

# Distribution-specific optimization
nix-foundry config optimize
```

### WSL
```bash
# Configure Windows path integration
nix-foundry config set windows-integration true

# Repair WSL environment
nix-foundry config repair
```

## Best Practices

### macOS
- Use multi-user installation for better security
- Configure shell integration before first use
- Enable binary cache for faster package installation

### Linux
- Check SELinux/AppArmor configuration
- Use distribution-specific optimizations
- Configure system limits for multi-user installations

### WSL
- Use WSL2 (not WSL1)
- Keep Windows paths in `/mnt`
- Use Linux filesystem for better performance

## Troubleshooting

### Platform-Specific Issues

#### macOS
- SIP interference
- Homebrew conflicts
- Architecture mismatches

#### Linux
- SELinux/AppArmor blocks
- Distribution-specific package conflicts
- systemd service issues

#### WSL
- Path conversion issues
- Windows filesystem performance
- Memory/disk space limitations

## Migration

### Between Platforms
- Export package list: `nix-foundry export packages`
- Transfer configuration: `nix-foundry config export`
- Import on new system: `nix-foundry config import`

### Version Updates
- Automatic updates: `nix-foundry update`
- Manual updates: Download new binary
- Configuration migration: Automatic
