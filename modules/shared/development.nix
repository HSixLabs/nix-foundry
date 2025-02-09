{ config, pkgs, lib, ... }:

{
  # System-level development packages
  environment.systemPackages = with pkgs; [
    git
    gnumake
    gcc
    nodejs
    python3
  ];

  # System-level development settings
  programs = {
    gnupg.agent = {
      enable = true;
      enableSSHSupport = true;
    };
  };
} 