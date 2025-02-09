{ config, pkgs, lib, users, hm, ... }:

{
  imports = [
    ./shell.nix
    ./git.nix
    ./vscode.nix
  ];

  home = {
    username = users.username;
    homeDirectory = users.homeDirectory;
    stateVersion = "23.11";
    
    sessionVariables = {
      EDITOR = "nvim";
      VISUAL = "code";
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