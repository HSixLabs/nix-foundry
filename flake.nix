{
  description = "Nix Configuration";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nix-darwin = {
      url = "github:LnL7/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixgl = {
      url = "github:guibou/nixGL";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    powerlevel10k = {
      url = "github:romkatv/powerlevel10k";
      flake = false;
    };
  };

  outputs = { nixpkgs, home-manager, nix-darwin, nixgl, ... }:
    let
      systems = [
        "aarch64-darwin"
        "x86_64-darwin"
        "aarch64-linux"
        "x86_64-linux"
        "x86_64-windows"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems (system: import nixpkgs {
        inherit system;
        config = {
          allowUnfree = true;
          allowUnsupportedSystem = true;
        };
      });

      mkUsersModule = { hostName, system }: {
        _module.args.users = import ./users.nix { inherit hostName system; };
      };

      defaultSystem = "aarch64-darwin";
    in
    {
      homeConfigurations."default" = let system = nixpkgs.system or defaultSystem; in 
        home-manager.lib.homeManagerConfiguration {
          pkgs = forAllSystems.${system};
          modules = [ 
            (mkUsersModule {
              hostName = builtins.getEnv "HOST";
              inherit system;
            })
            ./home/default.nix 
          ];
        };

      darwinConfigurations."default" = let 
        system = nixpkgs.system or defaultSystem;
        users = import ./users.nix { 
          hostName = builtins.getEnv "HOST";
          inherit system;
        };
      in 
        nix-darwin.lib.darwinSystem {
          inherit system;
          modules = [
            ./modules/darwin
            home-manager.darwinModules.home-manager
            (mkUsersModule {
              hostName = builtins.getEnv "HOST";
              inherit system;
            })
            {
              home-manager.useGlobalPkgs = true;
              home-manager.useUserPackages = true;
              home-manager.users.${users.username} = {
                imports = [
                  ./home/default.nix
                  ./home/darwin.nix
                ];
              };
            }
          ];
        };
    };
}