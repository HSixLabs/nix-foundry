# Contributing to nix-configs

Thanks for your interest in my Nix configuration! While this is my personal setup, I welcome contributions that might help others using it as a reference or template.

## Commit Message Guidelines

I use [Conventional Commits](https://www.conventionalcommits.org/) for automated semantic versioning. Your commit messages should follow this format:

```none
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature (minor version bump)
- `fix`: A bug fix (patch version bump)
- `docs`: Documentation changes
- `style`: Changes that don't affect code meaning
- `refactor`: Code changes that neither fix bugs nor add features
- `perf`: Performance improvements
- `test`: Adding or fixing tests
- `chore`: Changes to build process or auxiliary tools
- `ci`: Changes to CI configuration files and scripts

### Breaking Changes

Breaking changes can be indicated in two ways:

1. Adding a `!` after the type/scope: `feat!: remove support for X`
2. Adding `BREAKING CHANGE:` in the commit footer

Examples:

```bash
# Feature (minor version bump)
feat: add support for NixOS

# Bug fix (patch version bump)
fix: correct homebrew installation

# Breaking change (major version bump)
feat!: remove support for x86 windows
# or
feat: remove support for x86 windows

BREAKING CHANGE: x86 windows support has been removed
```

## Development Workflow

1. Clone the repository
2. Set up pre-commit hooks:

   ```bash
   # Install pre-commit
   nix-shell -p pre-commit nixpkgs-fmt

   # Install the hooks
   pre-commit install -t pre-commit -t commit-msg
   ```

3. Create a feature branch:

   ```bash
   git checkout -b feat/my-new-feature
   # or
   git checkout -b fix/bug-description
   ```

4. Make your changes
5. Stage your changes:

   ```bash
   git add .
   ```

   The pre-commit hooks will automatically:
   - Check for common issues (trailing whitespace, YAML format, etc.)
   - Format Nix files
   - Verify flake builds
   - Validate commit messages follow conventional commits format

6. Commit your changes (commitizen will help format your message)
7. Push your changes
8. Create a Pull Request

## Testing

Before submitting a PR:

1. Ensure the code builds:

   ```bash
   nix flake check
   ```

2. Test your changes on your platform:

   ```bash
   nix build .#homeConfigurations.$USER-$(uname -m)-$(uname -s | tr '[:upper:]' '[:lower:]').activationPackage
   ```

## Release Process

The release process is automated based on conventional commits:

- Breaking changes trigger a major version bump
- New features trigger a minor version bump
- Bug fixes and other changes trigger a patch version bump

When changes are merged to `main`:

1. Version is evaluated based on commit messages
2. Changelog is generated automatically
3. New release is created with generated notes
