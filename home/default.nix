{ config, pkgs, lib, users, hm, ... }:

{
  imports = [
    ./zsh.nix
    ./zsh-exports.nix
    ./zsh-aliases.nix
    ./zsh-vim-mode.nix
    ./git.nix
    ./vscode.nix
    ./darwin.nix
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

  programs.home-manager.enable = true;

  programs = {
    direnv = {
      enable = true;
      nix-direnv.enable = true;
    };
  };
} 