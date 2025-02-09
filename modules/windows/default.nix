{ config, pkgs, lib, users, ... }:

{
  imports = [
    ../shared/base.nix
  ];

  # Windows-specific settings
  home.sessionVariables = {
    MSYSTEM = "MINGW64";
    MSYS2_PATH_TYPE = "inherit";
  };

  # Windows-specific packages
  home.packages = with pkgs; [
    windows.mingw
    windows.msys2
  ];

  # Windows path handling
  home.sessionPath = [
    "$HOME/AppData/Local/Programs/Microsoft VS Code/bin"
    "$HOME/AppData/Local/Programs/Git/bin"
  ];

  # Windows-specific program configurations
  programs = {
    git = {
      enable = true;
      extraConfig = {
        core.autocrlf = "true";
        core.symlinks = "true";
      };
    };
  };
} 