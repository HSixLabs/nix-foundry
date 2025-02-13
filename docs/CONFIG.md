# Configuration Guide

## Basic Configuration

```yaml
# ~/.config/nix-foundry/config.yaml
shell:
  type: zsh
  plugins:
    - zsh-autosuggestions

editor:
  type: neovim

packages:
  - git
  - ripgrep
```

## Common Patterns

### Personal Setup
```yaml
shell:
  aliases:
    k: kubectl
  env:
    EDITOR: nvim

packages:
  - fd
  - bat
  - fzf
```

### Team Setup
```yaml
# .nix-foundry.yaml
required:
  - golang
  - docker
tools:
  go:
    - gopls
    - delve
```

For detailed configuration options, see [Configuration Reference](CONFIG-REFERENCE.md).

## Quick Reference

```yaml
# ~/.config/nix-foundry/config.yaml
shell:
  type: zsh
  plugins:
    - zsh-autosuggestions
  aliases:
    k: kubectl
    g: git

editor:
  type: neovim
  plugins:
    - telescope.nvim
    - lsp-zero.nvim

packages:
  - ripgrep      # Fast search
  - fd           # Better find
  - bat          # Better cat
  - git          # Version control
```

## Common Configurations

### 1. Development Tools
```yaml
development:
  languages:
    go:
      version: "1.21"
      tools: [gopls, delve]
    node:
      version: "18"
      packages: [typescript, eslint]
```

### 2. Environment Variables
```yaml
environment:
  personal:
    EDITOR: nvim
    GOPATH: "$HOME/go"
  team:           # Automatically merged with team settings
    NODE_ENV: development
```

## Team Integration

1. **Package Priority**
   - Team requirements always installed
   - Personal packages preserved when non-conflicting
   - Conflicts favor team settings

2. **Configuration Merging**
   - Shell preferences preserved
   - Team tools automatically added
   - Environment variables merged

## Best Practices

1. **Keep It Simple**
   - Start minimal, add as needed
   - Document custom settings
   - Use version control

2. **Performance**
   - Choose lightweight alternatives
   - Remove unused packages
   - Use conditional loading

Need more details? Check:
- [Full Reference](CONFIG-REFERENCE.md)
- [Team Guide](TEAM.md)
- [FAQ](FAQ.md)
