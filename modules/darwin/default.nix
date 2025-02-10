{ config, pkgs, lib, users, hostName, ... }:

{
  imports = [
    ./core.nix
    ./fonts.nix
    ./homebrew.nix
  ];

  # Fix for build user group ID mismatch
  ids.gids.nixbld = lib.mkForce 350;
  nix.settings.trusted-users = [ users.username "root" "@admin" "@wheel" ];

  # Darwin-specific environment settings
  environment = {
    systemPath = [
      "/opt/homebrew/bin"
      "/opt/homebrew/sbin"
      "/usr/local/bin"
      "/usr/bin"
      "/usr/sbin"
      "/bin"
      "/sbin"
    ];
    pathsToLink = [ "/Applications" ];
  };

  # User configuration
  users.users = {
    ${users.username} = {
      name = users.username;
      home = users.homeDirectory;
      shell = pkgs.zsh;
    };
  };

  # Security settings
  security.pam.enableSudoTouchIdAuth = true;

  # Used for backwards compatibility
  system.stateVersion = 4;
}
