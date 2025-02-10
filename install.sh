#!/usr/bin/env bash

setup_ssl_certs() {
  # Create certificates directory with proper permissions
  sudo mkdir -p /etc/ssl/certs
  sudo chmod 755 /etc/ssl/certs
  
  # Export and set up certificates
  echo "Setting up SSL certificates..."
  sudo security find-certificate -a -p /System/Library/Keychains/SystemRootCertificates.keychain > /tmp/certs.pem
  sudo security find-certificate -a -p /Library/Keychains/System.keychain >> /tmp/certs.pem
  sudo mv /tmp/certs.pem /etc/ssl/certs/ca-certificates.crt
  sudo chmod 644 /etc/ssl/certs/ca-certificates.crt
  
  # Verify certificates were created
  if [ ! -f "/etc/ssl/certs/ca-certificates.crt" ]; then
    echo "Error: Failed to create SSL certificates"
    exit 1
  fi
  
  # Set environment variables
  export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
  export NIX_SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
  export CURL_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt
}

detect_platform() {
  local os=$(uname -s)
  local arch=$(uname -m)
  
  case "$os" in
    Darwin*)
      case "$arch" in
        arm64) echo "aarch64-darwin" ;;
        x86_64) echo "x86_64-darwin" ;;
      esac
      ;;
    Linux*)
      case "$arch" in
        aarch64|arm64) 
          echo "aarch64-linux"
          ;;
        x86_64)
          if grep -q microsoft /proc/version; then
            echo "x86_64-linux-wsl"
          else
            echo "x86_64-linux"
          fi
          ;;
      esac
      ;;
    MINGW*|MSYS*|CYGWIN*)
      echo "x86_64-windows"
      ;;
    *)
      echo "unknown"
      exit 1
      ;;
  esac
}

fetch_file() {
  local path="$1"
  local output="$2"
  local repo="shawnkhoffman/nix-configs"
  local branch="main"
  
  mkdir -p "$(dirname "$output")"
  
  if [ -f "$path" ]; then
    cp "$path" "$output"
    return 0
  fi
  
  if [ -z "$GITHUB_TOKEN" ]; then
    echo "Error: GITHUB_TOKEN environment variable is required for remote installation"
    exit 1
  fi
  
  echo "Fetching $path..."
  local response=$(curl -fsSL \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3.raw" \
    -H "Cache-Control: no-cache, no-store, must-revalidate" \
    -H "Pragma: no-cache" \
    -H "Expires: 0" \
    -w "%{http_code}" \
    "https://api.github.com/repos/${repo}/contents/${path}?ref=${branch}" \
    -o "$output" 2>/dev/null)
    
  if [ "$response" != "200" ]; then
    echo "Error: Failed to fetch ${path} (HTTP ${response})"
    return 1
  fi
  
  if [ ! -s "$output" ]; then
    echo "Error: Empty file received for ${path}"
    return 1
  fi
  
  echo "Successfully fetched ${path}"
  return 0
}

setup_config_dir() {
  # Create necessary directories
  mkdir -p "$CONFIG_DIR"/{home,modules/{darwin,nixos,shared/programs}}

  # Core files
  fetch_file "flake.nix" "$CONFIG_DIR/flake.nix"
  fetch_file "users.nix" "$CONFIG_DIR/users.nix"

  # Home manager configurations
  for file in default.nix git.nix vscode.nix; do
    fetch_file "home/$file" "$CONFIG_DIR/home/$file"
  done

  # Shell configurations - platform aware
  case "$PLATFORM" in
    x86_64-windows)
      # Windows uses PowerShell by default
      fetch_file "home/pwsh.nix" "$CONFIG_DIR/home/pwsh.nix"
      mkdir -p "$APPDATA/PowerShell"
      ;;
    *)
      # Unix-like systems use ZSH
      fetch_file "home/zsh.nix" "$CONFIG_DIR/home/zsh.nix"
      
      # Create ZSH config directory based on platform
      case "$PLATFORM" in
        *-darwin)
          mkdir -p "$HOME/.config/zsh/conf.d"
          ;;
        *-linux*)
          mkdir -p "$HOME/.config/zsh/conf.d"
          ;;
      esac
      ;;
  esac

  # Shared modules (all platforms)
  fetch_file "modules/shared/base.nix" "$CONFIG_DIR/modules/shared/base.nix"
  fetch_file "modules/shared/programs/nix.nix" "$CONFIG_DIR/modules/shared/programs/nix.nix"

  # Platform-specific modules
  case "$PLATFORM" in
    *-darwin)
      fetch_file "modules/darwin/default.nix" "$CONFIG_DIR/modules/darwin/default.nix"
      fetch_file "modules/darwin/core.nix" "$CONFIG_DIR/modules/darwin/core.nix"
      fetch_file "modules/darwin/fonts.nix" "$CONFIG_DIR/modules/darwin/fonts.nix"
      ;;
    *-linux*)
      fetch_file "modules/nixos/default.nix" "$CONFIG_DIR/modules/nixos/default.nix"
      fetch_file "modules/nixos/hardware.nix" "$CONFIG_DIR/modules/nixos/hardware.nix"
      fetch_file "modules/nixos/network.nix" "$CONFIG_DIR/modules/nixos/network.nix"
      ;;
    x86_64-windows)
      # Windows-specific modules if needed
      ;;
  esac
}

# Add Windows-specific setup
setup_windows() {
  # Create symbolic links for Windows config files
  if [ "$PLATFORM" = "x86_64-windows" ]; then
    APPDATA="${HOME}/AppData/Roaming"
    mkdir -p "$APPDATA"
    
    # Setup Windows-specific paths
    if [ ! -d "$HOME/.config" ]; then
      mkdir -p "$HOME/.config"
    fi
  fi
}

main() {
  PLATFORM=$(detect_platform)
  HOST=$(hostname)

  # Always set both USER and USERNAME
  USER=$(whoami)
  USERNAME=$USER
  export USER USERNAME HOST PLATFORM

  # Setup SSL certificates first for Darwin
  if [[ "$PLATFORM" == *"-darwin" ]]; then
    setup_ssl_certs
  fi

  # Platform-specific configurations
  case "$PLATFORM" in
    *-darwin)
      CONFIG_DIR="$HOME/.config/nix-configs"
      ;;
    *-linux*)
      CONFIG_DIR="$HOME/.config/nix-configs"
      ;;
    x86_64-windows)
      CONFIG_DIR="$APPDATA/nix-configs"
      setup_windows
      ;;
    *)
      echo "Unsupported platform: $PLATFORM"
      exit 1
      ;;
  esac
  
  mkdir -p "$CONFIG_DIR"

  # Setup configuration directory and fetch files
  setup_config_dir

  # Install home-manager if not present
  if ! command -v home-manager >/dev/null 2>&1; then
    nix-channel --add https://github.com/nix-community/home-manager/archive/master.tar.gz home-manager
    nix-channel --update
    nix-shell '<home-manager>' -A install
  fi

  # Set NIX_USER for more reliable nix evaluation
  NIX_USER=$(whoami)
  export NIX_USER HOST

  # Build and activate with extraSpecialArgs
  nix run home-manager/master -- switch \
    --flake "$CONFIG_DIR#default" \
    --extra-experimental-features "nix-command flakes" \
    --impure
}

main "$@"

cat > ~/.config/zsh/conf.d/exports.zsh << 'EOL'
# XDG paths
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_DATA_HOME="$HOME/.local/share"
export XDG_CACHE_HOME="$HOME/.cache"

# Path
export PATH="$HOME/.local/bin:$PATH"

# Editor
export EDITOR="nvim"
export VISUAL="code"
export TERMINAL="alacritty"
export BROWSER="brave"
export MANPAGER='nvim +Man!'
export MANWIDTH=999
export PATH=$HOME/.cargo/bin:$PATH
export PATH=$HOME/.local/share/go/bin:$PATH
export GOPATH=$HOME/.local/share/go
export PATH=$HOME/.fnm:$PATH
export XDG_CURRENT_DESKTOP="Wayland"
EOL

cat > ~/.config/zsh/conf.d/vim-mode.zsh << 'EOL'
# Enable vi mode
bindkey -v
export KEYTIMEOUT=1

# Use vim keys in tab complete menu
bindkey -M menuselect 'h' vi-backward-char
bindkey -M menuselect 'k' vi-up-line-or-history
bindkey -M menuselect 'l' vi-forward-char
bindkey -M menuselect 'j' vi-down-line-or-history

# Change cursor shape for different vi modes
function zle-keymap-select {
  if [[ ${KEYMAP} == vicmd ]] ||
     [[ $1 = 'block' ]]; then
    echo -ne '\e[1 q'
  elif [[ ${KEYMAP} == main ]] ||
       [[ ${KEYMAP} == viins ]] ||
       [[ ${KEYMAP} = "" ]] ||
       [[ $1 = 'beam' ]]; then
    echo -ne '\e[5 q'
  fi
}
zle -N zle-keymap-select
EOL

cat > ~/.config/zsh/conf.d/aliases.zsh << 'EOL'
# Your aliases here
alias ls='ls --color=auto'
alias ll='ls -la'
alias vim='nvim'
# ... add your other aliases
EOL