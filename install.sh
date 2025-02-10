#!/usr/bin/env bash

REPO_URL=${NIX_CONFIG_REPO:-"https://github.com/shawnkhoffman/nix-configs.git"}

cleanup_nix() {
  echo "Cleaning up existing Nix installation..."
  
  if [ "$(uname -s)" = "Darwin" ]; then
    echo "Stopping Nix daemons..."
    # Stop services first
    sudo launchctl bootout system/org.nixos.nix-daemon 2>/dev/null || true
    sudo launchctl bootout system/org.nixos.darwin-store 2>/dev/null || true
    sudo rm -f /Library/LaunchDaemons/org.nixos.nix-daemon.plist
    sudo rm -f /Library/LaunchDaemons/org.nixos.darwin-store.plist
    
    echo "Waiting for services to stop..."
    sleep 5
    
    # More specific process killing with retries
    echo "Checking for Nix processes..."
    for attempt in {1..3}; do
      echo "Attempt $attempt to stop Nix processes..."
      
      # Kill specific processes in order
      for proc in "nix-daemon" "nix-store" "/nix/store.*"; do
        if pgrep -f "$proc" > /dev/null; then
          echo "Stopping $proc processes..."
          sudo pkill -15 -f "$proc" 2>/dev/null || true
          sleep 2
          sudo pkill -9 -f "$proc" 2>/dev/null || true
        fi
      done
      
      # Check if any critical Nix processes are still running
      if ! (pgrep -f "nix-daemon" > /dev/null || pgrep -f "nix-store" > /dev/null); then
        echo "Successfully stopped Nix processes"
        break
      fi
      
      if [ $attempt -eq 3 ]; then
        echo "Failed to stop all Nix processes after 3 attempts."
        echo "Please try manually running: sudo pkill -9 -f nix-daemon"
        exit 1
      fi
      
      sleep 5
    done
    
    echo "Waiting for processes to fully stop..."
    sleep 10
    
    # Handle volume unmounting
    if mount | grep -q "on /nix ("; then
      echo "Unmounting Nix volume..."
      
      # Try graceful unmount first
      if sudo diskutil unmount /nix 2>/dev/null; then
        echo "Successfully unmounted Nix volume"
      else
        echo "Graceful unmount failed, waiting..."
        sleep 10
        
        # Try force unmount
        if sudo diskutil unmount force /nix 2>/dev/null; then
          echo "Successfully force-unmounted Nix volume"
        else
          echo "Force unmount failed. Please restart your computer and try again."
          exit 1
        fi
      fi
      
      # Extra wait after unmounting
      sleep 10
    fi
    
    # Only try to remove volume if unmount was successful
    if ! mount | grep -q "on /nix ("; then
      if diskutil list | grep -q "Nix Store"; then
        echo "Removing Nix Store volume..."
        nix_volume=$(diskutil list | grep "Nix Store" | awk '{print $NF}')
        sudo diskutil apfs deleteVolume "$nix_volume" 2>/dev/null || true
        sleep 10
      fi
    fi
  fi
  
  # Restore backup files if they exist
  for file in bashrc zshrc bash.bashrc; do
    if [ -f "/etc/${file}.backup-before-nix" ]; then
      sudo mv "/etc/${file}.backup-before-nix" "/etc/${file}"
    fi
  done
  
  # Remove Nix files with proper permissions
  if [ -d "/nix" ]; then
    sudo rm -rf /nix/* 2>/dev/null || true
  fi
  
  # Clean up SSL certificates and Nix-related files
  sudo rm -rf "/etc/nix" "/var/root/.nix-*" "/var/root/.local/state/nix" "/var/root/.cache/nix"
  rm -rf "$HOME/.nix-*" "$HOME/.local/state/nix" "$HOME/.cache/nix"
  sudo rm -f /etc/ssl/certs/ca-certificates.crt
  
  # macOS-specific cleanup
  if [ "$(uname -s)" = "Darwin" ]; then
    if [ -f /etc/synthetic.conf ]; then
      sudo sed -i '' '/^nix/d' /etc/synthetic.conf
    fi
    sudo rm -f /etc/fstab
  fi
}

detect_os() {
  case "$(uname -s)" in
    Darwin*)
      if [ "$(uname -m)" = "arm64" ]; then
        echo "darwin-arm64"
      else
        echo "darwin-intel"
      fi
      ;;
    Linux*)
      if grep -q microsoft /proc/version; then
        echo "wsl"
      elif [ -f "/etc/cachyos-release" ]; then
        echo "cachyos"
      else
        echo "unknown-linux"
      fi
      ;;
    *)
      echo "unknown"
      ;;
  esac
}

main() {
  OS=$(detect_os)
  
  # Get the hostname and username
  HOSTNAME=$(hostname -s)
  USERNAME=$(whoami)
  
  # Export HOSTNAME and USERNAME for Nix evaluation
  export HOSTNAME="${HOSTNAME}"
  export USERNAME="${USERNAME}"
  
  case "$OS" in
    darwin*)
      cleanup_nix
      
      # Install Nix
      echo "Installing Nix..."
      sh <(curl -L https://nixos.org/nix/install)
      
      # Source nix environment
      . /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
      
      # Enable flakes
      mkdir -p ~/.config/nix
      echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf
      
      # Set up initial SSL certificates for building
      echo "Setting up initial SSL certificates..."
      sudo mkdir -p /etc/ssl/certs
      sudo security find-certificate -a -p /System/Library/Keychains/SystemRootCertificates.keychain > /tmp/ca-certificates.crt
      sudo mv /tmp/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
      sudo chmod 644 /etc/ssl/certs/ca-certificates.crt
      
      # Clone or update the configuration repository
      CONFIG_DIR="$HOME/.config/nix-configs"
      
      if [ ! -d "$CONFIG_DIR" ]; then
        echo "Cloning configuration repository..."
        git clone "$REPO_URL" "$CONFIG_DIR"
      else
        echo "Updating configuration repository..."
        (cd "$CONFIG_DIR" && git pull)
      fi
      
      # Source nix environment again to ensure it's available
      . /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh
      
      # Debug information
      echo "Hostname: ${HOSTNAME}"
      echo "Available configurations:"
      CURL_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt nix flake show "$CONFIG_DIR"
      
      # Backup existing VS Code settings if they exist
      echo "Backing up existing configuration files..."
      VSCODE_SETTINGS="$HOME/Library/Application Support/Code/User/settings.json"
      if [ -f "$VSCODE_SETTINGS" ]; then
        mv "$VSCODE_SETTINGS" "$VSCODE_SETTINGS.backup"
      fi
      
      # Install Powerlevel10k theme from nix flake instead of git clone
      echo "Installing Powerlevel10k theme..."
      mkdir -p ~/.config/zsh
      
      # After installing Nix and before building configuration
      echo "Starting Nix daemon..."
      sudo launchctl load /Library/LaunchDaemons/org.nixos.nix-daemon.plist
      sleep 5  # Give the daemon time to start
      
      # Then continue with the build
      echo "Building configuration for ${HOSTNAME}..."
      HOSTNAME="${HOSTNAME}" CURL_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt nix build "$CONFIG_DIR#darwinConfigurations.${HOSTNAME}.system" --show-trace
      
      # Move certificates out of the way before activation
      echo "Moving certificates for activation..."
      sudo mv /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt.before-nix-darwin
      
      # Then activate it
      echo "Activating system..."
      ./result/sw/bin/darwin-rebuild switch --flake "$CONFIG_DIR#${HOSTNAME}"
      ;;
    wsl)
      # Similar changes for wsl if needed...
      ;;
    *)
      echo "Unsupported operating system"
      exit 1
      ;;
  esac
}

main "$@"