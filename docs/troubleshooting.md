# Troubleshooting Guide

Common issues and their solutions when using Nix Foundry.

## Installation Issues

### Permission Denied

**Symptom**: Error when running `nix-foundry install`
```
Error: permission denied: /nix
```

**Solutions**:
1. For single-user installation:
```bash
sudo chown -R $(whoami) /nix
```

2. For multi-user installation:
```bash
sudo nix-foundry install --multi-user
```

### SSL Certificate Issues (macOS)

**Symptom**: SSL certificate validation errors
```
Error: unable to verify SSL certificates
```

**Solution**:
```bash
nix-foundry config set platform.darwin.sslCerts true
nix-foundry config apply
```

### WSL Path Issues

**Symptom**: Windows path conversion errors
```
Error: invalid path conversion
```

**Solution**:
```bash
nix-foundry config set platform.wsl.windowsIntegration true
nix-foundry config repair
```

## Configuration Issues

### Invalid Configuration

**Symptom**: Configuration validation errors
```
Error: invalid configuration at line X
```

**Solution**:
1. Validate configuration:
```bash
nix-foundry config validate
```

2. Reset to defaults:
```bash
nix-foundry config reset
```

### Shell Integration Failed

**Symptom**: Shell not properly configured
```
Error: shell integration not found
```

**Solution**:
```bash
nix-foundry setup --shell $(basename $SHELL) --force
```

## Package Management Issues

### Package Installation Failed

**Symptom**: Package fails to install
```
Error: failed to install package X
```

**Solutions**:
1. Update channels:
```bash
nix-foundry update --channels
```

2. Clear cache:
```bash
nix-foundry clean --cache
```

### Conflicting Packages

**Symptom**: Package conflicts during installation
```
Error: conflicting packages found
```

**Solution**:
```bash
nix-foundry packages list --conflicts
nix-foundry packages remove [conflicting-package]
```

## Platform-Specific Issues

### macOS

#### Homebrew Conflicts

**Symptom**: Conflicts between Nix and Homebrew
```
Error: package already installed via Homebrew
```

**Solution**:
```bash
nix-foundry config set platform.darwin.brewIntegration false
nix-foundry config apply
```

#### Apple Silicon Issues

**Symptom**: Architecture compatibility issues
```
Error: incompatible architecture
```

**Solution**:
```bash
nix-foundry config set platform.darwin.architecture arm64
nix-foundry config apply
```

### Linux

#### SELinux Blocking

**Symptom**: SELinux prevents operations
```
Error: SELinux blocked operation
```

**Solution**:
```bash
nix-foundry config set platform.linux.selinux true
nix-foundry repair --selinux
```

#### systemd Service Issues

**Symptom**: Daemon service not starting
```
Error: failed to start nix-daemon.service
```

**Solution**:
```bash
sudo nix-foundry repair --services
```

### WSL

#### Windows Filesystem Issues

**Symptom**: Slow operations on Windows filesystem
```
Warning: slow filesystem operations detected
```

**Solution**:
```bash
nix-foundry config set platform.wsl.mountPoint /nix
nix-foundry repair --filesystem
```

## Performance Issues

### Slow Package Operations

**Symptom**: Package operations are unusually slow

**Solutions**:
1. Configure binary cache:
```bash
nix-foundry config set nix.substituters[+] https://cache.nixos.org
```

2. Optimize storage:
```bash
nix-foundry clean --optimize
```

### High Disk Usage

**Symptom**: `/nix` store consuming too much space

**Solutions**:
1. Clean old generations:
```bash
nix-foundry clean --older-than 30d
```

2. Remove unused packages:
```bash
nix-foundry clean --unused
```

## Diagnostic Tools

### System Check

Run comprehensive diagnostics:
```bash
nix-foundry doctor --verbose
```

### Log Analysis

View system logs:
```bash
nix-foundry logs --level error
```

Export logs for support:
```bash
nix-foundry logs export --last 24h
```

### Configuration Validation

Validate current configuration:
```bash
nix-foundry config validate --strict
```

## Recovery Options

### Emergency Recovery

1. Reset to default state:
```bash
nix-foundry reset --all
```

2. Repair installation:
```bash
nix-foundry repair --full
```

### Backup and Restore

1. Backup current state:
```bash
nix-foundry backup create
```

2. Restore from backup:
```bash
nix-foundry backup restore [backup-id]
```

## Getting Support

1. Generate support bundle:
```bash
nix-foundry support-bundle create
```

2. Check system status:
```bash
nix-foundry status --all
```

3. Report an issue:
```bash
nix-foundry issue report
```

## Common Error Codes

- `NF001`: Permission error
- `NF002`: Configuration error
- `NF003`: Network error
- `NF004`: Platform compatibility error
- `NF005`: Package management error
- `NF006`: Shell integration error
- `NF007`: System resource error
- `NF008`: Platform-specific error

## Best Practices

1. **Regular Maintenance**
   - Run `nix-foundry doctor` weekly
   - Clean unused packages monthly
   - Update regularly

2. **Configuration Management**
   - Keep backups of working configurations
   - Use version control for configurations
   - Document custom changes

3. **System Resources**
   - Monitor disk usage
   - Configure appropriate resource limits
   - Use platform-specific optimizations
