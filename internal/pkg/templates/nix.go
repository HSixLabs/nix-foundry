package templates

const (
	// FlakeTemplate is the default template for flake.nix
	FlakeTemplate = `{
  description = "nix-foundry managed environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, home-manager, ... }: {
    defaultPackage.x86_64-linux = home-manager.defaultPackage.x86_64-linux;
    defaultPackage.x86_64-darwin = home-manager.defaultPackage.x86_64-darwin;

    homeConfigurations = {
      current = home-manager.lib.homeManagerConfiguration {
        configuration = ./home.nix;
        system = builtins.currentSystem;
        homeDirectory = builtins.getEnv "HOME";
        username = builtins.getEnv "USER";
      };
    };
  };
}`

	// HomeManagerTemplate is the default template for home.nix
	HomeManagerTemplate = `{ config, pkgs, ... }:

{
  # Home Manager needs a bit of information about you and the paths it should manage
  home.username = builtins.getEnv "USER";
  home.homeDirectory = builtins.getEnv "HOME";

  # Basic configuration
  home.stateVersion = "23.11";
  programs.home-manager.enable = true;

  # Let Home Manager install and manage itself
  programs.home-manager.enable = true;

  # Packages to install
  home.packages = with pkgs; [
    # Add your packages here
  ];

  # Environment variables
  home.sessionVariables = {
    EDITOR = "vim";
  };
}`
)
