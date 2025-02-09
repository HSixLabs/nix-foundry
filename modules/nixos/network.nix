{ config, pkgs, lib, ... }:

{
  networking = {
    networkmanager.enable = true;
    firewall = {
      enable = true;
      allowPing = true;
    };
  };
} 