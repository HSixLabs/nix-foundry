# Configuration Reference

## Quick Reference Table

| Setting | Description | Example |
|---------|-------------|---------|
| `shell.type` | Shell executable | `"zsh"` |
| `shell.plugins` | Shell plugins | `["zsh-autosuggestions"]` |
| `editor.type` | Editor choice | `"neovim"` |
| `packages` | Tool list | `["ripgrep", "git"]` |

## Personal Configuration

```yaml
# ~/.config/nix-foundry/config.yaml
version: "1.0"

shell:
  type: "zsh"
  plugins:
    - zsh-syntax-highlighting
  aliases:
    k: kubectl

editor:
  type: "neovim"
  plugins:
    - telescope.nvim
    - lsp-zero.nvim

packages:
  - ripgrep    # Fast search
  - fd         # Better find
  - git        # Version control
```

## Team Configuration

```yaml
# project/.nix-foundry.yaml
name: "project-name"
version: "1.0"

packages:
  required:
    - golang
    - docker
  recommended:
    - gopls
    - delve

environment:
  GO111MODULE: "on"
  NODE_ENV: "development"
```

## Configuration Rules

### 1. Package Management
- Team required → Always installed
- Personal packages → Preserved if no conflicts
- Version conflicts → Team version used

### 2. Environment Variables
- Team vars override personal
- PATH entries merged
- Shell vars respect personal config

### 3. Tool Settings
- Team tools take precedence
- Personal preferences kept for other tools
- Conflicts clearly reported

## Common Patterns

### Development Setup
```yaml
development:
  languages:
    go:
      version: "1.21"
      tools: [gopls, delve]
    node:
      version: "18"
      packages: [typescript]
```

### Environment Management
```yaml
environments:
  default:
    EDITOR: "nvim"
    PATH: ["$HOME/bin", "$HOME/.local/bin"]
  docker:
    DOCKER_BUILDKIT: "1"
```

Need help? See:
- [Getting Started](GETTING-STARTED.md)
- [Troubleshooting](TROUBLESHOOTING.md)
- [FAQ](FAQ.md)
