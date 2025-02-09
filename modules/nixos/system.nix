{ config, pkgs, lib, ... }:

{
  # Boot configuration
  boot = {
    loader = {
      systemd-boot.enable = true;
      efi.canTouchEfiVariables = true;
    };
  };

  # Networking
  networking = {
    networkmanager.enable = true;
    firewall.enable = true;
  };

  # Time zone and locale
  time.timeZone = "America/Chicago";
  i18n.defaultLocale = "en_US.UTF-8";

  # Sound
  sound.enable = true;
  hardware.pulseaudio.enable = true;

  # X11 and desktop
  services = {
    xserver = {
      enable = true;
      layout = "us";
      libinput.enable = true;
      displayManager.gdm.enable = true;
      desktopManager.gnome.enable = true;
    };
  };

  # User account
  users.users.shawnhoffman = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" "docker" ];
    shell = pkgs.zsh;
  };
} 