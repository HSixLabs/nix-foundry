{ config, pkgs, lib, ... }:

{
  # Enable ZSH system-wide
  programs.zsh.enable = true;

  # Make ZSH available system-wide
  environment.systemPackages = with pkgs; [
    zsh
  ];

  # Set ZSH as default shell for new users
  users.defaultUserShell = pkgs.zsh;
} 