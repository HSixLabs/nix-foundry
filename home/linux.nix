{ config, pkgs, lib, ... }:

{
  imports = [ ./default.nix ];

  home = {
    packages = with pkgs; [
      xclip
      gnome.gnome-tweaks
    ];
  };

  programs = {
    alacritty.enable = true;
  };

  services = {
    gpg-agent.enable = true;
  };
} 