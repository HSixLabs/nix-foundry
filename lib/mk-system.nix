{ type, inputs, system, hostName, homeModules ? [], ... }:

let
  pkgs = import inputs.nixpkgs {
    inherit system;
    config.allowUnfree = true;
  };
  
  users = import ../users.nix { inherit hostName system; };

  darwinConfiguration = inputs.darwin.lib.darwinSystem {
    inherit system;
    modules = [
      ../modules/darwin
      ../modules/shared/base.nix
      inputs.home-manager.darwinModules.home-manager
      {
        home-manager = {
          useGlobalPkgs = true;
          useUserPackages = true;
          users.${users.username} = {
            imports = homeModules;
            home = {
              sessionVariables = {
                ZDOTDIR = "${users.homeDirectory}/.config/zsh";
                ZSH_CACHE_DIR = "${users.homeDirectory}/.cache/zsh";
              };
            };
          };
        };
      }
    ];
  };

  homeConfiguration = inputs.home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = homeModules;
  };
in
if type == "darwin" then darwinConfiguration else homeConfiguration
