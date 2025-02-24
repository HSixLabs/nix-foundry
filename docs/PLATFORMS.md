# Platform Support

Nix Foundry supports multiple platforms with platform-specific optimizations.

## Linux

### Supported Distributions

- Ubuntu 20.04+
- Debian 10+
- Fedora 34+
- CentOS/RHEL 8+
- Arch Linux
- NixOS

### Installation Modes

- Multi-user
- Single-user

### Requirements

- x86_64 or aarch64 architecture
- systemd for multi-user mode
- sudo access for multi-user mode

## macOS

### Supported Versions

- macOS 11 (Big Sur) and newer
- Intel and Apple Silicon support

### Installation Mode

- Multi-user mode only (required)
- Admin privileges required

### Requirements

- Command Line Tools for Xcode
- Rosetta 2 (for Intel packages on Apple Silicon)
- System Integrity Protection (SIP) enabled

## Windows (WSL2)

### Requirements

- Windows 10 version 2004+ or Windows 11
- WSL2 enabled
- Ubuntu 20.04+ (recommended)
- Windows Terminal (recommended)

### Installation Mode

- Single-user mode recommended
- Multi-user mode available but requires additional setup

### Considerations

- Use Linux filesystem for best performance
- Some GUI applications require WSLg
- Network access follows WSL2 networking rules

## Common Requirements

All platforms require:

- Internet connection
- Bash or compatible shell
- curl or wget
- Git for development packages

## Platform-Specific Features

### Linux
- Native container support
- Full systemd integration
- Direct hardware access

### macOS
- Native ARM package support
- Automatic Rosetta 2 integration
- Seamless GUI application support

### WSL2
- Windows path integration
- Automatic port forwarding
- WSLg support for GUI applications

## Best Practices

1. Use recommended installation mode for your platform
2. Keep system packages updated
3. Follow platform-specific security guidelines
4. Use native packages when available
5. Consider filesystem locations for optimal performance
