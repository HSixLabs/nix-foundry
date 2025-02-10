#!/usr/bin/env bash

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
  curl -fsSL \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3.raw" \
    -H "Cache-Control: no-cache" \
    -v \
    "https://api.github.com/repos/${repo}/contents/${path}?ref=${branch}&$(date +%s)" \
    > "$output" || {
    echo "Error: Failed to fetch ${path}"
    return 1
  }
}

setup_config_dir() {
  # Create necessary directories
  mkdir -p "$CONFIG_DIR"/{home,modules/{darwin,nixos,shared/programs}}

  # Core files
  fetch_file "flake.nix" "$CONFIG_DIR/flake.nix"
  fetch_file "users.nix" "$CONFIG_DIR/users.nix"

  # Home manager configurations
  for file in default.nix git.nix shell.nix vscode.nix; do
    fetch_file "home/$file" "$CONFIG_DIR/home/$file"
  done

  # Shared modules
  fetch_file "modules/shared/base.nix" "$CONFIG_DIR/modules/shared/base.nix"
  fetch_file "modules/shared/programs/nix.nix" "$CONFIG_DIR/modules/shared/programs/nix.nix"
  fetch_file "modules/shared/programs/shell.nix" "$CONFIG_DIR/modules/shared/programs/shell.nix"
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

# Setup SSL certificates for macOS
if [[ "$PLATFORM" == *"-darwin" ]]; then
  # Create certificates directory with proper permissions
  sudo mkdir -p /etc/ssl/certs
  sudo chmod 755 /etc/ssl/certs
  
  # Export and set up certificates
  sudo security find-certificate -a -p /System/Library/Keychains/SystemRootCertificates.keychain > /tmp/certs.pem
  sudo security find-certificate -a -p /Library/Keychains/System.keychain >> /tmp/certs.pem
  sudo mv /tmp/certs.pem /etc/ssl/certs/ca-certificates.crt
  sudo chmod 644 /etc/ssl/certs/ca-certificates.crt
  
  # Set environment variables
  export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
  export NIX_SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
fi

main() {
  PLATFORM=$(detect_platform)
  HOST=$(hostname)

  # Always set both USER and USERNAME
  USER=$(whoami)
  USERNAME=$USER
  export USER USERNAME HOST PLATFORM

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