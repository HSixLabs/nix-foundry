package config

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func Generate(configDir string, cfg *NixConfig) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate flake.nix
	if err := generateFlake(configDir, cfg); err != nil {
		return fmt.Errorf("failed to generate flake.nix: %w", err)
	}

	// Generate home configuration
	if err := generateHome(configDir, cfg); err != nil {
		return fmt.Errorf("failed to generate home config: %w", err)
	}

	return nil
}

const flakeTemplate = `{
  description = "Nix development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    {{if eq .Platform.OS "darwin"}}
    nix-darwin = {
      url = "github:LnL7/nix-darwin";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    {{end}}

    {{if eq .Platform.OS "linux"}}
    nixgl = {
      url = "github:guibou/nixGL";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    {{end}}
  };

  outputs = { self, nixpkgs, home-manager
    {{if eq .Platform.OS "darwin"}}, nix-darwin{{end}}
    {{if eq .Platform.OS "linux"}}, nixgl{{end}}
  }: let
    # Support all major platforms
    supportedSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];

    # Helper function to generate attrs for each system
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

    # System-specific package sets
    pkgsForSystem = system: import nixpkgs {
      inherit system;
      config = {
        allowUnfree = true;
        allowUnsupportedSystem = builtins.getEnv "NIX_ALLOW_UNSUPPORTED" == "1";
      };
    };

    # Current system configuration
    system = "{{if eq .Platform.Arch "arm64"}}aarch64{{else}}{{.Platform.Arch}}{{end}}-{{.Platform.OS}}";
  in {
    # Home Manager configuration for the current system
    homeConfigurations.default =
      home-manager.lib.homeManagerConfiguration {
        pkgs = pkgsForSystem system;
        modules = [
          ./home.nix
          {
            home.sessionVariables = {
              {{if eq .Platform.OS "darwin"}}
              PATH = "$HOME/bin:/usr/local/bin:$PATH";
              {{else}}
              PATH = "$HOME/.local/bin:$HOME/bin:$PATH";
              {{end}}
            };
          }
          (if builtins.pathExists ./custom.nix then ./custom.nix else {})
        ];
      };

    {{if eq .Platform.OS "darwin"}}
    # macOS-specific configuration
    darwinConfigurations.default = nix-darwin.lib.darwinSystem {
      inherit system;
      modules = [
        home-manager.darwinModules.home-manager
        {
          home-manager.useGlobalPkgs = true;
          home-manager.useUserPackages = true;
          home-manager.users.${builtins.getEnv "USER"}.imports = [ ./home.nix ];
        }
      ];
    };
    {{end}}

    # Make the configuration available for all supported systems
    packages = forAllSystems (system: {
      default = pkgsForSystem system;
    });

    # Add flake utilities
    lib = {
      inherit pkgsForSystem;
      inherit supportedSystems;
    };
  };
}`

func generateFlake(configDir string, cfg *NixConfig) error {
	f, err := os.Create(filepath.Join(configDir, "flake.nix"))
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("flake").Parse(flakeTemplate))
	return tmpl.Execute(f, cfg)
}
