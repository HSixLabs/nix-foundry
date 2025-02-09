{ config, pkgs, lib, ... }:

{
  # Boot configuration
  boot = {
    loader = {
      systemd-boot.enable = true;
      efi.canTouchEfiVariables = true;
    };
    tmp.cleanOnBoot = true;
  };

  # Networking configuration
  networking = {
    networkmanager.enable = true;
    firewall = {
      enable = true;
      allowPing = false;
    };
  };

  # Time and locale settings
  time.timeZone = "America/New_York";
  i18n.defaultLocale = "en_US.UTF-8";

  # Sound configuration
  sound.enable = true;
  hardware.pulseaudio.enable = true;

  # Core system programs
  programs = {
    zsh = {
      enable = true;
      enableCompletion = true;
      syntaxHighlighting.enable = true;
      autosuggestions.enable = true;
    };
    gnupg.agent = {
      enable = true;
      enableSSHSupport = true;
    };
  };

  # Core system services
  services = {
    # Enable DBus
    dbus.enable = true;

    # Enable CUPS for printing
    printing.enable = true;

    # Enable the OpenSSH daemon
    openssh = {
      enable = true;
      settings = {
        PermitRootLogin = "no";
        PasswordAuthentication = false;
      };
    };

    # Enable Flatpak support
    flatpak.enable = true;
  };

  # Security settings
  security = {
    rtkit.enable = true;
    polkit.enable = true;
    sudo = {
      enable = true;
      wheelNeedsPassword = true;
    };
  };

  # Font configuration
  fonts = {
    packages = with pkgs; [
      noto-fonts
      noto-fonts-cjk
      noto-fonts-emoji
      (nerdfonts.override { fonts = [ "Meslo" "Hack" ]; })
    ];
    fontconfig = {
      defaultFonts = {
        serif = [ "Noto Serif" ];
        sansSerif = [ "Noto Sans" ];
        monospace = [ "Hack Nerd Font" ];
      };
    };
  };
}
