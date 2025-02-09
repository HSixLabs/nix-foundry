{ config, pkgs, lib, ... }:

{
  imports = [ ./default.nix ];

  home = {
    # Linux-specific packages
    packages = with pkgs; [
      xclip
      gnome.gnome-tweaks
    ];
  };

  # Linux-specific program configurations
  programs = {
    alacritty.enable = true;
  };

  # Linux-specific services
  services = {
    gpg-agent.enable = true;
  };
} 