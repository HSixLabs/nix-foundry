{ config, pkgs, lib, users, hostName, ... }:

{
  imports = [
    ./core.nix
    ./fonts.nix
    ./homebrew.nix
  ];

  ids.gids.nixbld = lib.mkForce 350;
  nix.settings.trusted-users = [ users.username "root" "@admin" "@wheel" ];

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

  users.users = {
    ${users.username} = {
      name = users.username;
      home = users.homeDirectory;
      shell = pkgs.zsh;
    };
  };

  security.pam.enableSudoTouchIdAuth = true;

  system.stateVersion = 4;
}
