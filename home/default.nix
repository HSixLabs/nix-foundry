{ config, pkgs, lib, users, hm, ... }:

{
  imports = [
    ./zsh.nix
    ./zsh-exports.nix
    ./zsh-aliases.nix
    ./zsh-vim-mode.nix
    ./git.nix
    ./vscode.nix
  ];

  home = {
    username = users.username;
    homeDirectory = lib.mkForce users.homeDirectory;
    stateVersion = "23.11";
    
    sessionVariables = {
      EDITOR = "nvim";
      VISUAL = "code";
      SHELL = "${pkgs.zsh}/bin/zsh";
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