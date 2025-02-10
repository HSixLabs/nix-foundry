#!/usr/bin/env bash

REPO_URL=${NIX_CONFIG_REPO:-"https://github.com/shawnkhoffman/nix-configs"}
RAW_URL="https://raw.githubusercontent.com/shawnkhoffman/nix-configs/main" 
GITHUB_TOKEN=${GITHUB_TOKEN:-""}
AUTH_HEADER=${GITHUB_TOKEN:+"Authorization: token ${GITHUB_TOKEN}"}

# Default values
DEFAULT_HOSTNAME=$(hostname -s)
DEFAULT_USERNAME=$(whoami)

# Parse command line arguments or use prompts
HOSTNAME=${1:-$(read -p "Enter hostname [$DEFAULT_HOSTNAME]: " hn; echo ${hn:-$DEFAULT_HOSTNAME})}
USERNAME=${2:-$(read -p "Enter username [$DEFAULT_USERNAME]: " un; echo ${un:-$DEFAULT_USERNAME})}

cleanup_nix() {
  echo "Cleaning up existing Nix installation..."
  
  if [ "$(uname -s)" = "Darwin" ]; then
    echo "Stopping Nix daemons..."
    sudo launchctl bootout system/org.nixos.nix-daemon 2>/dev/null || true
    sudo rm -f /Library/LaunchDaemons/org.nixos.nix-daemon.plist
    
    echo "Waiting for services to stop..."
    sleep 5

    echo "Checking for Nix processes..."
    for attempt in {1..3}; do
      echo "Attempt $attempt to stop Nix processes..."
      sudo pkill -15 -f "nix-daemon|nix-store" 2>/dev/null || true
      sleep 2
      sudo pkill -9 -f "nix-daemon|nix-store" 2>/dev/null || true

      if ! (pgrep -f "nix-daemon" > /dev/null || pgrep -f "nix-store" > /dev/null); then
        echo "Successfully stopped Nix processes"
        break
      fi

      if [ $attempt -eq 3 ]; then
        echo "Failed to stop all Nix processes after 3 attempts."
        exit 1
      fi
      
      sleep 5
    done
  fi
}

detect_os() {
  case "$(uname -s)" in
    Darwin*) echo "darwin" ;;
    Linux*) 
      if grep -q microsoft /proc/version; then
        echo "wsl"
      else
        echo "linux"
      fi
      ;;
    *) echo "unknown" ;;
  esac
}

generate_host_config() {
  local hostname="$1"
  local username="$2"
  local os_type="$3"
  local system_type="$4"

  local config_dir="$HOME/.config/nix-configs"
  local template="$config_dir/hosts.template.nix"
  local output="$config_dir/hosts.nix"

  # Create hosts.nix using the template
  cat > "$output" <<EOF
let
  hostTemplate = import ./hosts.template.nix { inherit (builtins) lib; };
in
  hostTemplate.mkHostConfig {
    hostname = "$hostname";
    username = "$username";
    system = "$system_type";
    type = "$os_type";
  }
EOF
}

fetch_file() {
  local path="$1"
  local output="$2"
  
  mkdir -p "$(dirname "$output")"
  if [ -n "$GITHUB_TOKEN" ]; then
    curl -H "$AUTH_HEADER" -L "${RAW_URL}/${path}" -o "$output"
  else
    curl -L "${RAW_URL}/${path}" -o "$output"
  fi
}

setup_config_dir() {
  CONFIG_DIR="$HOME/.config/nix-configs"
  mkdir -p "$CONFIG_DIR"/{modules/{darwin,nixos,shared},home,lib}

  # Fetch core files
  fetch_file "flake.nix" "$CONFIG_DIR/flake.nix"
  fetch_file "hosts.template.nix" "$CONFIG_DIR/hosts.template.nix"
  
  # Fetch modules
  for module in darwin/{default,fonts}.nix nixos/{default,hardware,network}.nix shared/{base,desktop}.nix; do
    fetch_file "modules/$module" "$CONFIG_DIR/modules/$module"
  done

  # Fetch home configurations
  for config in {default,darwin,linux,shell,git,vscode}.nix; do
    fetch_file "home/$config" "$CONFIG_DIR/home/$config"
  done
}

install_prerequisites() {
  OS=$(detect_os)
  if [ "$OS" = "darwin" ]; then
    echo "Installing nix-darwin prerequisites..."
    nix-build https://github.com/LnL7/nix-darwin/archive/master.tar.gz -A installer
    ./result/bin/darwin-installer
  fi
}

main() {
  OS=$(detect_os)
  USERNAME=$(whoami)
  export USERNAME

  CONFIG_DIR="$HOME/.config/nix-configs"
  mkdir -p "$CONFIG_DIR"

  # Setup configuration directory and fetch files
  setup_config_dir

  # Install home-manager if not present
  if ! command -v home-manager >/dev/null 2>&1; then
    nix-channel --add https://github.com/nix-community/home-manager/archive/master.tar.gz home-manager
    nix-channel --update
    nix-shell '<home-manager>' -A install
  fi

  # Build and activate
  home-manager switch --flake "$CONFIG_DIR#$USERNAME"
}

main "$@"