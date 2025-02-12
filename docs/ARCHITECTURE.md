# Architecture

nix-foundry is built on several key components that work together to provide a flexible, cross-platform development environment.

## Core Components

### 1. Platform Detection

The system uses a sophisticated platform detection system (see `lib/platform.nix`):

- Darwin (macOS) support for both Intel and Apple Silicon
- Linux support for x86_64 and ARM
- Windows support via WSL2
- Automatic path and shell configuration

### 2. User Management

Dynamic user detection across platforms:

- Environment-based user detection
- Platform-specific home directories
- Shell configuration adaptation
- Cross-platform path normalization

### 3. Package Management

Multi-layer package management strategy:

- Nix packages for core functionality
- Homebrew integration for macOS
- Platform-specific package handling
- Version pinning and reproducibility

### 4. Configuration System

Modular configuration architecture:

- Platform-specific modules (`modules/{darwin,nixos}`)
- Shared configurations (`modules/shared`)
- User-specific settings (`home/`)
- Local customization support

## System Flow

1. **Bootstrap Process**
   - Platform detection
   - SSL certificate setup (Darwin)
   - Package manager initialization
   - Configuration directory setup

2. **Configuration Application**
   - User environment setup
   - Package installation
   - System preferences configuration
   - Shell environment configuration

3. **Update Management**
   - Automated version detection
   - Change detection and backup
   - Selective updates
   - Configuration preservation

## Quality Assurance

### Testing Infrastructure

- Multi-platform CI pipeline
- Pre-commit hook validation
- Build verification
- Platform-specific tests

### Security Measures

- SSL certificate management
- Secure package sources
- Permission management
- Environment isolation

## Extension Points

### Custom Configuration

- Local shell configuration (`~/.zshrc.local`)
- Platform-specific overrides
- Package customization
- Development tool configuration

### Development Workflow

- Pre-commit hooks
- Conventional commits
- Automated releases
- Documentation generation

## Directory Structure

```shell
.
├── docs/               # Documentation
├── home/              # User environment configurations
├── lib/               # Platform and utility functions
├── modules/           # System configurations
│   ├── darwin/        # macOS-specific modules
│   ├── nixos/         # Linux-specific modules
│   └── shared/        # Cross-platform modules
└── tests/             # Platform and integration tests
```

## Future Architecture

### Planned Improvements

- Windows native support enhancement
- Container integration
- Remote development support
- Multi-user environment improvements

### Compatibility Goals

- Maintain cross-platform consistency
- Ensure backward compatibility
- Support future Nix features
- Expand platform support
