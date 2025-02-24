# Changelog

All notable changes to this project will be documented in this file. See [semantic-release](https://github.com/semantic-release/semantic-release) for commit guidelines.

<!-- semantic-release will add new releases here automatically -->

## Commit Types

### Major Version Bumps (Breaking Changes)
Breaking changes can be indicated in two ways:

1. Using `!` after the type:
```
feat!: change configuration file format
fix!: remove deprecated API
```

2. Using `BREAKING CHANGE` in the commit body:
```
feat: change configuration file format

BREAKING CHANGE: The configuration file format has changed from YAML to TOML.
Old config files will need to be migrated.
```

### Minor Version Bumps
- **feat**: A new feature
- **feat(api)**: API-related features

### Patch Version Bumps
- **fix**: A bug fix
- **perf**: A performance improvement
- **docs**: Documentation changes
- **refactor**: Code refactoring
- **build**: Build system changes
- **deps**: Dependency updates
- **go**: Go-specific changes
- **go(mod)**: Go module updates

### No Version Bump
- **style**: Code style changes
- **test**: Test changes
- **ci**: CI changes

## Installation

Using curl:
```bash
curl -L https://nixfoundry.org/install | bash
