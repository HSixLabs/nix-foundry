# Getting Started with Nix Foundry

This guide will help you get up and running with Nix Foundry quickly.

## Quick Installation

### macOS and Linux

```bash
curl -L https://github.com/shawnkhoffman/nix-foundry/releases/latest/download/nix-foundry-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o /usr/local/bin/nix-foundry
chmod +x /usr/local/bin/nix-foundry
```

### WSL

```bash
curl -L https://github.com/shawnkhoffman/nix-foundry/releases/latest/download/nix-foundry-linux-amd64 -o /usr/local/bin/nix-foundry
chmod +x /usr/local/bin/nix-foundry
```

## First Steps

1. Initialize Nix Foundry:
```bash
nix-foundry init
```

2. Install Nix:
```bash
# Single-user installation
nix-foundry install

# Multi-user installation (macOS/Linux only)
sudo nix-foundry install --multi-user
```

3. Verify installation:
```bash
nix-foundry --version
nix-env --version
```

## Basic Usage

### Managing Packages

Install a package:
```bash
nix-foundry packages install git
```

List installed packages:
```bash
nix-foundry packages list
```

Search for packages:
```bash
nix-foundry packages search nodejs
```

Remove a package:
```bash
nix-foundry packages remove git
```

### Managing Scripts

Add a development setup script:
```bash
nix-foundry script add setup-dev.sh --name "Setup Dev" --desc "Setup development environment"
```

List available scripts:
```bash
nix-foundry script list
```

Run a script:
```bash
nix-foundry script run "Setup Dev"
```

## Configuration

### View Current Configuration

```bash
nix-foundry config list
```

### Basic Settings

Enable automatic updates:
```bash
nix-foundry config set settings.autoUpdate true
```

Change log level:
```bash
nix-foundry config set settings.logLevel debug
```

### Package Management

Add core packages:
```bash
nix-foundry config set nix.packages.core[+] git
nix-foundry config set nix.packages.core[+] curl
```

### Shell Configuration

Set default shell:
```bash
nix-foundry config set nix.shell /bin/zsh
```

## Common Tasks

### Development Environment Setup

1. Create a development configuration:
```bash
cat > dev-config.yaml << EOF
version: "1.0"
kind: config
metadata:
  name: dev-environment
nix:
  packages:
    core:
      - git
      - nodejs
      - python3
    optional:
      - docker
      - vscode
EOF
```

2. Import the configuration:
```bash
nix-foundry config import dev-config.yaml
```

3. Apply the configuration:
```bash
nix-foundry config apply
```

### Platform-Specific Setup

#### macOS

Enable Homebrew integration:
```bash
nix-foundry config set platform.darwin.brewIntegration true
```

#### Linux

Configure SELinux support:
```bash
nix-foundry config set platform.linux.selinux true
```

#### WSL

Setup Windows integration:
```bash
nix-foundry config set platform.wsl.windowsIntegration true
```

## Maintenance

### Updates

Update Nix Foundry:
```bash
nix-foundry update --self
```

Update Nix:
```bash
nix-foundry update --nix
```

### System Cleanup

Clean unused packages:
```bash
nix-foundry clean --all
```

### Health Check

Run system diagnostics:
```bash
nix-foundry doctor
```

## Best Practices

1. **Regular Updates**
   - Keep Nix Foundry updated
   - Regularly update installed packages
   - Monitor system health with `doctor` command

2. **Configuration Management**
   - Keep configuration in version control
   - Use environment-specific configurations
   - Regularly backup configurations

3. **Script Organization**
   - Use descriptive script names
   - Add proper descriptions
   - Include platform requirements

4. **Performance**
   - Clean unused packages regularly
   - Monitor disk usage
   - Use platform-specific optimizations

## Next Steps

- Read the full [Command Reference](commands.md)
- Learn about [Platform Support](platform-support.md)
- Explore [Configuration Options](configuration.md)
- Check [Troubleshooting](troubleshooting.md) for common issues

## Getting Help

- Run `nix-foundry help [command]` for detailed command information
- Check the [documentation](README.md) for comprehensive guides
- Visit the [GitHub repository](https://github.com/shawnkhoffman/nix-foundry) for issues and updates
- Join the community discussions in GitHub Discussions
