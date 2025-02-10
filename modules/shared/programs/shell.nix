{ config, pkgs, lib, ... }:

{
  programs.zsh.enable = true;

  environment.systemPackages = with pkgs; [
    zsh
  ];
} 