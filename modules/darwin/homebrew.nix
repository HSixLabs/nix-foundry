{ config, pkgs, lib, ... }:

{
  homebrew = {
    enable = true;
    onActivation = {
      autoUpdate = true;
      cleanup = "zap";
      upgrade = true;
    };
    
    # Generate Brewfile on changes
    global.brewfile = true;
    
    # Save Brewfile to a consistent location
    global.lockfiles = {
      enable = true;
      path = "~/.config/Brewfile";
    };

    taps = [
      "homebrew/bundle"
      "homebrew/cask"
      "homebrew/core"
      "homebrew/services"
      # Add other taps from apps.nix
    ];

    # Automatically track new installations
    global.autoUpdate = true;
  };

  # Add hook to backup Brewfile after changes
  system.activationScripts.postActivation.text = ''
    # Backup Brewfile if it exists
    if [ -f ~/.config/Brewfile ]; then
      cp ~/.config/Brewfile ~/.config/Brewfile.backup
    fi
    
    # Install packages from Brewfile
    if [ -f ~/.config/Brewfile ]; then
      brew bundle install --file=~/.config/Brewfile
    fi
  '';
} 