{ config, pkgs, ... }:

{
  programs.iterm2 = {
    enable = true;
    # Enable integration with nix-darwin
    enableIntegration = true;
  };
}
