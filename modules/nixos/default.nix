{ config, pkgs, lib, users, ... }:

{
  imports = [
    ./hardware.nix
    ./network.nix
  ];

  # Basic system configuration
  time.timeZone = "America/New_York";

  # User configuration
  users.users.${users.username} = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" "docker" ];
    shell = pkgs.zsh;
    home = "/home/${users.username}";
  };

  # Enable basic services
  services = {
    openssh.enable = true;
    printing.enable = true;
  };

  # System state version
  system.stateVersion = "23.11";
}
