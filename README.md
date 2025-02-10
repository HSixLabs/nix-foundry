# nix-configs

## Installation

Using curl:

```bash
curl -H "Authorization: token ${GITHUB_TOKEN}" -L https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash
```

Or using wget:

```bash
wget -qO- --header="Authorization: token ${GITHUB_TOKEN}" https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main/install.sh | bash
```

## Post-install Configuration

Configure p10k to use the conf.d directory:

```bash
POWERLEVEL9K_CONFIG_FILE="$HOME/.config/zsh/conf.d/p10k.zsh" p10k configure
```
