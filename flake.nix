{
  description = "Shawn's Nix Configs";

  inputs = {
    # Core
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    
    # System Management
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    darwin = {
      url = "github:lnl7/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    
    # Hardware Support
    nixos-hardware.url = "github:nixos/nixos-hardware";
    
    # Development
    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    
    # Utils
    flake-utils.url = "github:numtide/flake-utils";

    zsh-powerlevel10k = {
      url = "github:romkatv/powerlevel10k";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, home-manager, darwin, flake-utils, rust-overlay, nixos-hardware, zsh-powerlevel10k, ... } @ inputs:
    let
      hostName = builtins.getEnv "HOSTNAME";
      # Add debug trace
      _ = builtins.trace "Flake evaluation: HOSTNAME=${hostName}" null;
      
      users = import ./users.nix {
        inherit hostName;
      };
      inherit (nixpkgs) lib;
      hm = inputs.home-manager.lib;
      
      # Helper function to create Darwin system configuration
      mkDarwinSystem = hostname: 
        let
          # Add debug trace
          _ = builtins.trace "Creating Darwin system for hostname=${hostname}" null;
        in
        darwin.lib.darwinSystem {
          system = if builtins.currentSystem == "aarch64-darwin" 
                  then "aarch64-darwin" 
                  else "x86_64-darwin";
          specialArgs = { inherit users hostName lib hm; };
          modules = [
            ./modules/darwin
            ./modules/shared/base.nix
            home-manager.darwinModules.home-manager
            {
              networking.hostName = hostname;
              nix.settings = {
                trusted-users = [ users.username ];
                experimental-features = [ "nix-command" "flakes" ];
              };
              system.stateVersion = 4;
              nix.configureBuildUsers = true;
              home-manager = {
                useGlobalPkgs = true;
                useUserPackages = true;
                users.${users.username} = {
                  imports = [ ./home/darwin.nix ];
                  home = {
                    username = users.username;
                    homeDirectory = users.homeDirectory;
                    stateVersion = "23.11";
                  };
                };
              };
            }
          ];
        };

      # Helper function to create NixOS system configuration
      mkNixosSystem = { system, hostname }: nixpkgs.lib.nixosSystem {
        inherit system;
        specialArgs = { inherit users hostName lib hm inputs; };
        modules = [
          ./modules/nixos
          ./modules/shared/base.nix
          home-manager.nixosModules.home-manager
          {
            networking.hostName = hostname;
            nix.settings = {
              trusted-users = [ users.username ];
              experimental-features = [ "nix-command" "flakes" ];
            };
            system.stateVersion = "23.11";
            home-manager = {
              useGlobalPkgs = true;
              useUserPackages = true;
              users.${users.username} = {
                imports = [ ./home/linux.nix ];
                home = {
                  username = users.username;
                  homeDirectory = users.homeDirectory;
                  stateVersion = "23.11";
                };
              };
            };
          }
        ];
      };
    in
    {
      darwinConfigurations = {
        ${hostName} = mkDarwinSystem hostName;
      };

      nixosConfigurations = {
        ${hostName} = mkNixosSystem {
          system = "x86_64-linux";
          hostname = hostName;
        };
      };
      
      devShells = flake-utils.lib.eachDefaultSystem (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              nixfmt
              nil
              statix
            ];
          };
        }
      );
    };
}