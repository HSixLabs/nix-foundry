{ config, pkgs, lib, ... }:

{
  # System configuration
  system = {
    defaults = {
      dock = {
        autohide = false;
        mru-spaces = false;
        minimize-to-application = true;
        show-process-indicators = true;
        show-recents = false;
      };
      
      finder = {
        AppleShowAllExtensions = true;
        FXEnableExtensionChangeWarning = false;
        _FXShowPosixPathInTitle = true;
      };
      
      NSGlobalDomain = {
        AppleShowAllExtensions = true;
        InitialKeyRepeat = 15;
        KeyRepeat = 2;
        AppleShowScrollBars = "Always";
        NSDocumentSaveNewDocumentsToCloud = false;
        NSAutomaticCapitalizationEnabled = false;
        NSAutomaticDashSubstitutionEnabled = false;
        NSAutomaticPeriodSubstitutionEnabled = false;
        NSAutomaticQuoteSubstitutionEnabled = false;
        NSAutomaticSpellingCorrectionEnabled = false;
      };
    };

    keyboard = {
      enableKeyMapping = true;
      remapCapsLockToEscape = true;
    };

    activationScripts = {
      preActivation.text = ''
        printf "creating /run directory... "
        sudo mkdir -p /run
        sudo chown root:wheel /run
        sudo chmod 755 /run
        echo "done"
      '';
      postActivation.text = ''
        # Reload system configuration
        /System/Library/PrivateFrameworks/SystemAdministration.framework/Resources/activateSettings -u
      '';
    };
  };

  # Darwin-specific services
  services = {
    nix-daemon.enable = true;
  };

  # Core system programs
  programs = {
    zsh.enable = true;
    nix-index.enable = true;
  };

  # Allow unfree packages
  nixpkgs.config.allowUnfree = true;
}
