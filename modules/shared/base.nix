{ config, pkgs, lib, ... }:

{
  imports = [
    ./programs/nix.nix
    ./programs/shell.nix
  ];

  # Common system packages
  environment.systemPackages = with pkgs; [
    coreutils
    curl
    git
    vim
    wget
  ];

  # Allow unfree packages globally
  nixpkgs.config.allowUnfree = true;
} 