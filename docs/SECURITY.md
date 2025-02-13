# Security Policy

## Quick Reference

- Report vulnerabilities to: security@nix-foundry.dev
- Supported versions: 1.x.x
- Security updates: Monthly or as needed
- Emergency patches: Within 24 hours

## Security Features

### 1. Environment Isolation
```yaml
# Automatic isolation between:
- Personal and team environments
- Different project environments
- System and user space
```

### 2. Permission Management
```bash
# Check current permissions
nix-foundry doctor --security

# Review elevated operations
nix-foundry audit --privileges
```

### 3. Data Protection
```yaml
sensitive_data:
  storage: encrypted
  backup: encrypted
  transmission: TLS 1.3
```

## Reporting Vulnerabilities

1. **Contact**
   - Email: security@nix-foundry.dev
   - Subject: "SECURITY: Brief description"
   - Include reproduction steps

2. **Expected Response**
   - Initial response: 24 hours
   - Status update: 72 hours
   - Fix timeline: Based on severity

## Best Practices

### 1. Environment Security
- Keep nix-foundry updated
- Use environment isolation
- Enable audit logging

### 2. Configuration Safety
- Validate all configs
- Use minimal permissions
- Regular security scans

### 3. Team Security
- Review team configs
- Monitor environment changes
- Regular security updates

Need help? See:
- [Troubleshooting](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)
- [Team Guide](TEAM.md)
