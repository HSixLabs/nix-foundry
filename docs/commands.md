# Command Reference

Complete reference for all Nix Foundry CLI commands.

## Global Flags

```bash
--verbose, -v    Enable verbose output
--quiet, -q      Suppress all output except errors
--json           Output in JSON format
--config FILE    Use alternate config file
--no-color       Disable colored output
```

## Core Commands

### install

Install Nix package manager.

```bash
nix-foundry install [flags]

Flags:
--multi-user     Install in multi-user mode (requires sudo)
--no-daemon      Skip daemon installation in multi-user mode
--force          Force installation even if Nix exists
--darwin-ssl     Configure Darwin SSL certificates
```

### uninstall

Remove Nix installation.

```bash
nix-foundry uninstall [flags]

Flags:
--keep-store     Keep the Nix store intact
--keep-config    Preserve configuration files
--force          Force uninstallation
```

### update

Update Nix Foundry and Nix installation.

```bash
nix-foundry update [flags]

Flags:
--self           Update only Nix Foundry
--nix            Update only Nix
--channel NAME   Update to specific channel
```

## Configuration Commands

### config

Manage Nix Foundry configuration.

```bash
nix-foundry config SUBCOMMAND [flags]

Subcommands:
  set KEY VALUE    Set configuration value
  get KEY          Get configuration value
  list             List all configuration
  import FILE      Import configuration
  export FILE      Export configuration
  reset            Reset to defaults
```

### init

Initialize Nix Foundry configuration.

```bash
nix-foundry init [flags]

Flags:
--minimal         Minimal configuration
--with-examples   Include example configuration
--force          Overwrite existing configuration
```

### setup

Configure shell environment.

```bash
nix-foundry setup [flags]

Flags:
--shell SHELL    Specify shell (bash|zsh|fish)
--backup         Backup existing configuration
--no-modify      Print changes without applying
```

## Package Management

### packages

Manage Nix packages using nix-env.

```bash
nix-foundry packages SUBCOMMAND [flags]

Subcommands:
  install PKG     Install package
  remove PKG      Remove package
  list            List installed packages
  search QUERY    Search available packages
  update          Update package list
```

### script

Manage shell scripts.

```bash
nix-foundry script SUBCOMMAND [flags]

Subcommands:
  add FILE       Add script to configuration
  remove NAME    Remove script
  list           List configured scripts
  run NAME       Execute script
```

## Maintenance Commands

### clean

Clean up Nix store and temporary files.

```bash
nix-foundry clean [flags]

Flags:
--all            Remove all unused packages
--older-than N   Remove items older than N days
--dry-run        Show what would be removed
```

### repair

Repair Nix installation.

```bash
nix-foundry repair [flags]

Flags:
--check          Only check for issues
--permissions    Fix permissions only
--links          Fix symbolic links only
```

### doctor

Diagnose system issues.

```bash
nix-foundry doctor [flags]

Flags:
--fix            Attempt to fix issues
--verbose        Show detailed diagnostics
```

## Testing Commands

### test

Run system tests.

```bash
nix-foundry test [flags]

Flags:
--skip-platform     Skip platform tests
--skip-integration  Skip integration tests
--keep-temp        Keep temporary files
```

### test-summary

View test results.

```bash
nix-foundry test-summary [flags]

Flags:
--latest           Show latest test results
--since DURATION   Show tests newer than duration
--platform OS      Filter by platform
--type TYPE        Filter by test type
```

## Development Commands

### version

Show version information.

```bash
nix-foundry version [flags]

Flags:
--short          Print version number only
--json           Output in JSON format
```

### completion

Generate shell completion scripts.

```bash
nix-foundry completion SHELL

Arguments:
  SHELL           bash|zsh|fish|powershell
```

## Environment Variables

- `NIX_FOUNDRY_CONFIG`: Path to configuration file
- `NIX_FOUNDRY_LOG_LEVEL`: Log level (debug|info|warn|error)
- `NIX_FOUNDRY_NO_COLOR`: Disable colored output
- `NIX_FOUNDRY_HOME`: Override home directory
- `NIX_FOUNDRY_CACHE_DIR`: Override cache directory

## Exit Codes

- 0: Success
- 1: General error
- 2: Invalid usage
- 3: Configuration error
- 4: Permission error
- 5: Network error
- 6: Platform error
