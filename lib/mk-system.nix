{ type, inputs, system, hostName, homeModules ? [], ... }:

let
  pkgs = import inputs.nixpkgs {
    inherit system;
    config.allowUnfree = true;
  };
  
  users = import ../users.nix { inherit hostName; };

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
              username = users.username;
              homeDirectory = "/Users/${users.username}";
              sessionVariables = {
                ZDOTDIR = "/Users/${users.username}/.config/zsh";
                ZSH_CACHE_DIR = "/Users/${users.username}/.cache/zsh";
              };
            };
          };
        };
      }
    ];
  };

  homeConfiguration = inputs.home-manager.lib.homeManagerConfiguration {
    inherit pkgs;
    modules = homeModules ++ [
      {
        home = {
          username = users.username;
          homeDirectory = "/home/${users.username}";
        };
      }
    ];
  };
in
if type == "darwin" then darwinConfiguration else homeConfiguration
