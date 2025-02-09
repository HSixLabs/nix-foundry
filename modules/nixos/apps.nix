{ config, pkgs, lib, ... }:

{
  # System-wide applications
  environment.systemPackages = with pkgs; [
    # Desktop Environment
    gnome.gnome-tweaks
    gnome.dconf-editor
    
    # Development
    vscode
    docker
    
    # System Tools
    gparted
    htop
    
    # Media
    vlc
    spotify
    
    # Communication
    discord
    slack
    
    # Browsers
    firefox
    google-chrome
  ];

  # Enable flatpak support
  services.flatpak.enable = true;

  # Enable Docker
  virtualisation.docker = {
    enable = true;
    autoPrune = {
      enable = true;
      dates = "weekly";
    };
  };
}
