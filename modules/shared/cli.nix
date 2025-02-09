{ config, pkgs, lib, ... }:

{
  # CLI tools that should be available system-wide
  environment.systemPackages = with pkgs; [
    bat
    fd
    fzf
    htop
    jq
    ripgrep
    tree
  ];

  # Shell configuration
  programs = {
    bash.enable = true;
    zsh = {
      enable = true;
      enableCompletion = true;
    };
  };
}
