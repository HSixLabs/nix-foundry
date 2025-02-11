#!/usr/bin/env bash

setup_ssl_certs() {
  sudo mkdir -p /etc/ssl/certs
  sudo chmod 755 /etc/ssl/certs
  
  echo "Setting up SSL certificates..."
  sudo security find-certificate -a -p /System/Library/Keychains/SystemRootCertificates.keychain > /tmp/certs.pem
  sudo security find-certificate -a -p /Library/Keychains/System.keychain >> /tmp/certs.pem
  sudo mv /tmp/certs.pem /etc/ssl/certs/ca-certificates.crt
  sudo chmod 644 /etc/ssl/certs/ca-certificates.crt
  
  if [ ! -f "/etc/ssl/certs/ca-certificates.crt" ]; then
    echo "Error: Failed to create SSL certificates"
    exit 1
  fi
  
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
  mkdir -p "$CONFIG_DIR"/{home,modules/{darwin,nixos,shared/programs}}

  fetch_file "flake.nix" "$CONFIG_DIR/flake.nix"
  fetch_file "users.nix" "$CONFIG_DIR/users.nix"

  for file in default.nix git.nix vscode.nix zsh.nix zsh-exports.nix zsh-aliases.nix zsh-vim-mode.nix zsh-functions.nix darwin.nix; do
    if ! fetch_file "home/$file" "$CONFIG_DIR/home/$file"; then
      echo "Error: Failed to fetch home/$file"
      exit 1
    fi
  done

  case "$PLATFORM" in
    x86_64-windows)
      fetch_file "home/pwsh.nix" "$CONFIG_DIR/home/pwsh.nix"
      mkdir -p "$APPDATA/PowerShell"
      ;;
    *)
      mkdir -p "$HOME/.config/zsh/conf.d"
      rm -f "$HOME/.config/zsh/conf.d/vim-mode.zsh"
      ;;
  esac

  fetch_file "modules/shared/base.nix" "$CONFIG_DIR/modules/shared/base.nix"
  fetch_file "modules/shared/programs/nix.nix" "$CONFIG_DIR/modules/shared/programs/nix.nix"

  case "$PLATFORM" in
    *-darwin)
      fetch_file "modules/darwin/default.nix" "$CONFIG_DIR/modules/darwin/default.nix"
      fetch_file "modules/darwin/core.nix" "$CONFIG_DIR/modules/darwin/core.nix"
      fetch_file "modules/darwin/fonts.nix" "$CONFIG_DIR/modules/darwin/fonts.nix"
      fetch_file "modules/darwin/homebrew.nix" "$CONFIG_DIR/modules/darwin/homebrew.nix"
      ;;
    *-linux)
      fetch_file "modules/nixos/default.nix" "$CONFIG_DIR/modules/nixos/default.nix"
      fetch_file "modules/nixos/core.nix" "$CONFIG_DIR/modules/nixos/core.nix"
      ;;
  esac
}

setup_windows() {
  if [ "$PLATFORM" = "x86_64-windows" ]; then
    APPDATA="${HOME}/AppData/Roaming"
    mkdir -p "$APPDATA"
    
    if [ ! -d "$HOME/.config" ]; then
      mkdir -p "$HOME/.config"
    fi
  fi
}

setup_homebrew() {
  if [[ "$PLATFORM" != *"-darwin" ]]; then
    return 0
  fi

  local brew_path=""
  if [[ "$(uname -m)" == "arm64" ]]; then
    brew_path="/opt/homebrew/bin/brew"
  else
    brew_path="/usr/local/bin/brew"
  fi

  if command -v brew >/dev/null 2>&1; then
    echo "Homebrew is already installed and in PATH"
    return 0
  fi

  if [ -f "$brew_path" ]; then
    echo "Homebrew is installed but not in PATH, adding to PATH..."
    eval "$($brew_path shellenv)"
    return 0
  fi

  echo "Installing Homebrew..."
  NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  
  eval "$($brew_path shellenv)"

  mkdir -p "$HOME/.config/zsh/conf.d"
  cat > "$HOME/.config/zsh/conf.d/homebrew.zsh" << EOL
eval "\$($brew_path shellenv)"

function brew() {
  command brew "\$@"
  if [[ "\$1" == "install" ]] || [[ "\$1" == "uninstall" ]] || [[ "\$1" == "tap" ]] || [[ "\$1" == "untap" ]]; then
    command brew bundle dump --force --file="\$HOME/.config/Brewfile"
    if [ -f "\$HOME/.config/Brewfile" ]; then
      mkdir -p "\$HOME/.config/nix-configs"
      cp "\$HOME/.config/Brewfile" "\$HOME/.config/nix-configs/Brewfile"
    fi
  fi
}
EOL

  if ! [ -f "$HOME/.config/Brewfile" ]; then
    fetch_file "Brewfile" "$HOME/.config/Brewfile"
  fi
}

handle_update() {
  local config_dir="$1"
  
  if [ ! -d "$config_dir" ]; then
    echo "No existing configuration found at $config_dir"
    echo "Performing fresh installation..."
    return 0
  fi

  echo "Updating existing configuration at $config_dir"
  
  local backup_dir="$config_dir.backup-$(date +%Y%m%d-%H%M%S)"
  echo "Creating backup at $backup_dir"
  cp -r "$config_dir" "$backup_dir"
  
  setup_config_dir_update
  
  return 0
}

setup_config_dir_update() {
  mkdir -p "$CONFIG_DIR"/{home,modules/{darwin,nixos,shared/programs}}

  for file in "flake.nix" "users.nix"; do
    if [ ! -f "$CONFIG_DIR/$file" ] || ! has_local_changes "$CONFIG_DIR/$file"; then
      fetch_file "$file" "$CONFIG_DIR/$file"
    fi
  done
}

has_local_changes() {
  local file="$1"
  local temp_file="/tmp/$(basename "$file")"
  
  fetch_file "$(basename "$file")" "$temp_file"
  
  if [ ! -f "$file" ]; then
    rm -f "$temp_file"
    return 1
  fi
  
  if diff -q "$file" "$temp_file" >/dev/null 2>&1; then
    rm -f "$temp_file"
    return 1
  fi
  
  rm -f "$temp_file"
  return 0
}

handle_reinstall() {
  local config_dir="$1"
  
  if [ -d "$config_dir" ]; then
    echo "Existing configuration detected at $config_dir"
    echo "Removing existing configuration for clean reinstall..."
    rm -rf "$config_dir"
  fi
}

main() {
  if [ -z "$1" ]; then
    echo "Error: Operation type is required"
    echo "Usage: $0 <install|update|reinstall>"
    echo ""
    echo "Operations:"
    echo "  install   - Perform a fresh installation"
    echo "  update    - Update existing installation while preserving local changes"
    echo "  reinstall - Remove existing configuration and perform a clean reinstall"
    exit 1
  fi
  
  OPERATION="$1"
  
  case "$OPERATION" in
    install|update|reinstall)
      echo "Performing $OPERATION operation..."
      ;;
    *)
      echo "Invalid operation: $OPERATION"
      echo "Usage: $0 <install|update|reinstall>"
      echo ""
      echo "Operations:"
      echo "  install   - Perform a fresh installation"
      echo "  update    - Update existing installation while preserving local changes"
      echo "  reinstall - Remove existing configuration and perform a clean reinstall"
      exit 1
      ;;
  esac

  PLATFORM=$(detect_platform)
  HOST=$(hostname)

  USER=$(whoami)
  USERNAME=$USER
  export USER USERNAME HOST PLATFORM

  if [[ "$PLATFORM" == *"-darwin" ]]; then
    setup_ssl_certs
    setup_homebrew
  fi

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

  case "$OPERATION" in
    install)
      ;;
    update)
      handle_update "$CONFIG_DIR"
      ;;
    reinstall)
      handle_reinstall "$CONFIG_DIR"
      ;;
  esac
  
  if [ "$OPERATION" != "update" ]; then
    setup_config_dir
  fi

  if ! command -v home-manager >/dev/null 2>&1; then
    nix-channel --add https://github.com/nix-community/home-manager/archive/master.tar.gz home-manager
    nix-channel --update
    nix-shell '<home-manager>' -A install
  fi

  NIX_USER=$(whoami)
  export NIX_USER HOST

  nix run home-manager/master -- switch \
    --flake "$CONFIG_DIR#default" \
    --extra-experimental-features "nix-command flakes" \
    --impure

  echo "$OPERATION completed successfully!"
  echo "Note: You may need to restart your shell for all changes to take effect"
}

main "$@"