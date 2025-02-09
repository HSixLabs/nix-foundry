{ config, pkgs, lib, ... }:

{
  programs.zsh = {
    enable = true;
    enableCompletion = true;
    enableAutosuggestions = true;
    enableSyntaxHighlighting = true;
  };

  environment.systemPackages = with pkgs; [
    zsh
    zsh-autosuggestions
    zsh-syntax-highlighting
  ];
} 