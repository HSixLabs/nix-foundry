{
  description = "Nix Configuration";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, home-manager, ... }:
    let
      system = builtins.currentSystem;
      pkgs = nixpkgs.legacyPackages.${system};
      username = builtins.getEnv "USER";
      homeDirectory = if system == "aarch64-darwin" || system == "x86_64-darwin"
                     then "/Users/${username}"
                     else "/home/${username}";
    in {
      packages.${system}.default = home-manager.defaultPackage.${system};
      
      homeConfigurations.${username} = home-manager.lib.homeManagerConfiguration {
        inherit pkgs;
        extraSpecialArgs = {
          users = {
            inherit username homeDirectory;
          };
        };
        modules = [ ./home/default.nix ];
      };
    };
}