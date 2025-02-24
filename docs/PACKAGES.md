# Packages

Nix Foundry provides access to the entire Nixpkgs collection, with convenient package management and grouping features.

## Package Access

Nix Foundry can install any package available in the Nixpkgs repository. The examples below show common packages, but you're not limited to these - any package in Nixpkgs can be installed.

## Common Package Groups

These are some commonly used package groups, provided for convenience:

### Development
- `git`: Git version control
- `make`: GNU Make build tool
- `gcc`: GNU Compiler Collection
And many more development tools available in Nixpkgs.

### Web
- `nodejs`: Node.js runtime with npm
- `yarn`: Package manager
- `nginx`: Web server
Plus any other web development tools from Nixpkgs.

### Data
- `postgresql`: PostgreSQL database
- `redis`: Redis in-memory store
- `mongodb`: MongoDB database
And all other database systems available in Nixpkgs.

## Installation

Install any Nixpkgs package:
```bash
# Install any package from Nixpkgs
nix-foundry install <package-name>

# Examples
nix-foundry install nodejs
nix-foundry install python3
nix-foundry install vscode
```

## Configuration

Packages can be specified in your configuration:
```yaml
nix:
  packages:
    core:      # Required packages
      - git
      - curl
    optional:  # Additional packages
      - nodejs
      - any-nixpkgs-package
```

## Platform Support

Packages are automatically handled based on your platform:
- Linux: Native package names
- macOS: Intel and ARM packages automatically selected
- WSL2: Linux packages with Windows integration

## Language Support

Language packages include their essential tools. For example:
- `python3`: Python with pip
- `nodejs`: Node.js with npm
- `go`: Go with gopls
- `openjdk`: Java with Maven

But you can install any language or tool available in Nixpkgs.

## Finding Packages

Search for available packages:
```bash
# Search Nixpkgs
nix-foundry search <query>

# Show package details
nix-foundry show <package>
```

## Best Practices

1. Use core packages for essential tools
2. Document package dependencies
3. Use team configurations for shared dependencies
4. Consider platform compatibility
5. Pin package versions when needed

Remember: While we provide convenient package groups and examples, you can use any package from the Nixpkgs repository with Nix Foundry.
