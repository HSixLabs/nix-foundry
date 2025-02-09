{ config, pkgs, lib, ... }:

{
  imports = [
    ./base.nix
    ./packages.nix
    ./cli.nix
    ./development.nix
  ];
}
