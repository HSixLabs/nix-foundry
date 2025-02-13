# Frequently Asked Questions

## General

### What is nix-foundry?
nix-foundry is a dual-purpose environment manager that handles both personal development setups and team project environments using Nix and home-manager.

### How is it different from just using Nix?
nix-foundry adds team collaboration features, automatic conflict resolution, and seamless switching between personal and team environments while maintaining the reproducibility benefits of Nix.

## Installation

### Why do I need both Nix and home-manager?
- Nix provides the package management foundation
- home-manager manages user environment configuration
- Together they enable complete environment reproducibility

### Can I use nix-foundry without home-manager?
No, home-manager is required for managing personal configurations. It's automatically installed during setup.

## Configuration

### How are conflicts handled?
1. Team requirements take precedence for project tools
2. Personal preferences are preserved when non-conflicting
3. Conflicts are clearly reported with resolution options
4. Manual override options are available

### Can I have different configurations per project?
Yes! Each project can have its own `.nix-foundry.yaml` with specific requirements, while maintaining your personal preferences.

## Team Usage

### How do I share team configurations?
1. Commit `.nix-foundry.yaml` to your project repository
2. Team members run `nix-foundry project import`
3. Configuration is automatically merged with personal settings

### Can I temporarily disable team requirements?
Yes, use `nix-foundry switch personal` to temporarily use only personal settings.

## Troubleshooting

### Common Issues

#### "Configuration conflict detected"
- Review the conflict details with `nix-foundry doctor`
- Choose which settings to keep
- Update team or personal configuration accordingly

#### "Package not found"
- Verify package name in nixpkgs
- Check your Nix channel version
- Try `nix-foundry update` to refresh package lists

#### "Environment switch failed"
1. Check logs with `nix-foundry logs`
2. Verify configuration files
3. Run `nix-foundry doctor` for diagnostics
4. Try `nix-foundry rollback` if needed

### Platform-Specific

#### Linux
- SELinux conflicts: Temporarily disable or create appropriate policies
- Permission issues: Verify user permissions and group membership
- System packages: Use `nix-foundry doctor` to check dependencies

#### macOS
- Rosetta 2 needed for some packages on Apple Silicon
- SIP limitations: Some system modifications require SIP adjustment
- Homebrew conflicts: nix-foundry manages conflicts automatically

## Best Practices

### Personal Configuration
- Keep personal configurations minimal
- Use version control for dotfiles
- Regular backups with `nix-foundry backup create`

### Team Configuration
- Document required packages
- Use platform-agnostic tools when possible
- Regular testing across team member setups

### Development Workflow
- Test changes in isolation
- Use `nix-foundry doctor` regularly
- Keep environments updated
- Document platform-specific requirements

## Getting Help

### Where can I find more help?
1. Run `nix-foundry help [command]`
2. Check the [documentation](https://github.com/shawnkhoffman/nix-foundry/tree/main/docs)
3. Open an issue on GitHub
4. Join the community chat

### How do I report bugs?
1. Run `nix-foundry doctor` and save the output
2. Collect relevant logs with `nix-foundry logs`
3. Open an issue with:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details
