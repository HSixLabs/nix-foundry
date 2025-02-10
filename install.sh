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
  for file in default.nix git.nix vscode.nix zsh.nix zsh-exports.nix zsh-aliases.nix zsh-vim-mode.nix zsh-functions.nix; do
    if ! fetch_file "home/$file" "$CONFIG_DIR/home/$file"; then
      echo "Error: Failed to fetch home/$file"
      exit 1
    fi
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
      mkdir -p "$HOME/.config/zsh/conf.d"
      ;;
  esac

  # Shared modules
  fetch_file "modules/shared/base.nix" "$CONFIG_DIR/modules/shared/base.nix"
  fetch_file "modules/shared/programs/nix.nix" "$CONFIG_DIR/modules/shared/programs/nix.nix"

  # Platform-specific modules
  case "$PLATFORM" in
    *-darwin)
      fetch_file "modules/darwin/default.nix" "$CONFIG_DIR/modules/darwin/default.nix"
      fetch_file "modules/darwin/core.nix" "$CONFIG_DIR/modules/darwin/core.nix"
      fetch_file "modules/darwin/fonts.nix" "$CONFIG_DIR/modules/darwin/fonts.nix"
      ;;
    *-linux)
      fetch_file "modules/nixos/default.nix" "$CONFIG_DIR/modules/nixos/default.nix"
      fetch_file "modules/nixos/core.nix" "$CONFIG_DIR/modules/nixos/core.nix"
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

setup_homebrew() {
  # Only run Homebrew setup on Darwin systems
  if [[ "$PLATFORM" != *"-darwin" ]]; then
    return 0
  fi

  if ! command -v brew >/dev/null 2>&1; then
    echo "Installing Homebrew..."
    NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  fi

  # Ensure Homebrew is in PATH for both arm64 and x86_64
  local brew_path=""
  if [[ "$(uname -m)" == "arm64" ]]; then
    brew_path="/opt/homebrew/bin/brew"
  else
    brew_path="/usr/local/bin/brew"
  fi

  # Add Homebrew to PATH for current session
  eval "$($brew_path shellenv)"

  # Initialize Homebrew in shell config
  mkdir -p "$HOME/.config/zsh/conf.d"
  cat > "$HOME/.config/zsh/conf.d/homebrew.zsh" << EOL
# Initialize Homebrew
eval "\$($brew_path shellenv)"

# Auto-update Brewfile after package operations
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

  # Fetch existing Brewfile from repository if available
  if ! [ -f "$HOME/.config/Brewfile" ]; then
    fetch_file "Brewfile" "$HOME/.config/Brewfile"
  fi
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
  PLATFORM=$(detect_platform)
  HOST=$(hostname)

  # Always set both USER and USERNAME
  USER=$(whoami)
  USERNAME=$USER
  export USER USERNAME HOST PLATFORM

  # Setup SSL certificates and Homebrew for Darwin
  if [[ "$PLATFORM" == *"-darwin" ]]; then
    setup_ssl_certs
    setup_homebrew
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

  # Handle reinstallation if configuration already exists
  handle_reinstall "$CONFIG_DIR"
  
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

  echo "Installation completed successfully!"
  echo "Note: You may need to restart your shell for all changes to take effect"
}

main "$@"