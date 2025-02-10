{ config, pkgs, lib, users, hm, ... }:

{
  imports = [
    ./zsh.nix
    ./git.nix
    ./vscode.nix
    ./p10k.nix
  ];

  home = {
    username = users.username;
    homeDirectory = lib.mkForce users.homeDirectory;
    stateVersion = "23.11";
    
    sessionVariables = {
      EDITOR = "nvim";
      VISUAL = "code";
      ZDOTDIR = "$HOME/.config/zsh";
    };
  };

  # Let Home Manager install and manage itself
  programs.home-manager.enable = true;

  # Common program configurations
  programs = {
    direnv = {
      enable = true;
      nix-direnv.enable = true;
    };
  };
} 