# Platform-Specific Setup

nix-foundry provides optimized configurations for each supported platform.

## macOS (Darwin)

### macOS Prerequisites

- Xcode Command Line Tools
- Rosetta 2 (for Apple Silicon)

### macOS Features

- Homebrew integration
- System preferences configuration
- iTerm2 settings
- Font installation

### macOS Post-Install

The system will automatically:

- Configure SSL certificates
- Install Homebrew packages
- Apply system defaults

## Linux

### Linux Prerequisites

- systemd-based distribution
- curl or wget

### Linux Features

- Native package management
- System-wide configurations
- WSL2 support (when applicable)

### Linux Post-Install

- System configuration applied automatically
- Reboot may be required for kernel changes

## Windows (Experimental)

### Windows Prerequisites

- Windows 10/11
- WSL2 enabled
- Windows Terminal (recommended)

### Windows Features

- PowerShell configuration
- Windows Terminal settings
- Cross-platform compatibility

### Windows Known Limitations

- Some features require manual setup
- Performance may vary under WSL2

## Troubleshooting

### macOS Issues

1. **SSL Certificate Problems**

   ```bash
   # Regenerate certificates
   sudo rm /etc/ssl/certs/ca-certificates.crt
   curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-foundry/main/install.sh | bash -s -- reinstall
   ```

2. **Homebrew Integration Fails**

   - Ensure Xcode CLI tools are installed: `xcode-select --install`
   - Check Homebrew permissions: `sudo chown -R $(whoami) /opt/homebrew`

### Linux Issues

1. **Missing System Dependencies**

   ```bash
   # Ubuntu/Debian
   sudo apt update && sudo apt install curl git

   # Fedora
   sudo dnf install curl git
   ```

2. **Home Manager Conflicts**

   - Backup and remove existing config: `mv ~/.config/home-manager ~/.config/home-manager.bak`
   - Clear generation: `home-manager generations | head -1 | xargs home-manager remove-generations`

### Windows/WSL2 Issues

1. **WSL2 Installation**

   ```powershell
   # Run in PowerShell as Administrator
   wsl --install
   wsl --set-default-version 2
   ```

2. **Path Resolution Problems**
   - Check Windows/WSL path integration: `wsl.exe --status`
   - Ensure Windows Terminal uses correct shell: `"defaultProfile": "{WSL GUID}"`

### General Fixes

1. **Nix Store Corruption**

   ```bash
   nix-store --verify --check-contents
   nix-store --repair
   ```

2. **Configuration Not Applied**

   ```bash
   # Clear and rebuild
   rm -rf ~/.config/nix-configs
   curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-foundry/main/install.sh | bash -s -- reinstall
   ```
