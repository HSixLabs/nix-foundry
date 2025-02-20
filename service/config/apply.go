package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/shawnkhoffman/nix-foundry/pkg/validator"
)

type ApplyService struct {
	fs filesystem.FileSystem
}

func NewApplyService(fs filesystem.FileSystem) *ApplyService {
	return &ApplyService{fs: fs}
}

func (a *ApplyService) ActivateConfig(configPath string) error {
	cfg, nixConfig, err := a.generateNixConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to generate Nix configuration: %w", err)
	}

	configDir := filepath.Dir(configPath)
	nixDir := filepath.Join(configDir, "nix")

	if err := a.fs.CreateDir(nixDir); err != nil {
		return fmt.Errorf("failed to create nix config directory: %w", err)
	}

	shell := cfg.Nix.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
	}

	shellFile := filepath.Join(nixDir, "flake.nix")
	if err := a.fs.WriteFile(shellFile, []byte(nixConfig), 0644); err != nil {
		return fmt.Errorf("failed to write flake.nix: %w", err)
	}

	fmt.Println("\n🚀 Applying configuration...")

	cmd := exec.Command("nix", "develop", nixDir, "--ignore-environment", "--command", shell, "-i")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"NIX_FOUNDRY_ACTIVE=1",
		fmt.Sprintf("NIX_FOUNDRY_CONFIG=%s", configPath),
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run development shell: %w", err)
	}

	return nil
}

func (a *ApplyService) generateNixFromConfig(cfg schema.Config) string {
	nixConfig := "{\n"
	nixConfig += "  description = \"Nix Foundry development environment\";\n\n"
	nixConfig += "  inputs.nixpkgs.url = \"github:NixOS/nixpkgs/nixpkgs-unstable\";\n\n"
	nixConfig += "  outputs = { self, nixpkgs }:\n"
	nixConfig += "    let\n"
	nixConfig += "      supportedSystems = [\"x86_64-linux\" \"aarch64-linux\" \"x86_64-darwin\" \"aarch64-darwin\"];\n"
	nixConfig += "      forEachSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);\n"
	nixConfig += "    in {\n"
	nixConfig += "      devShells = forEachSystem (system: {\n"
	nixConfig += "        default = let\n"
	nixConfig += "          pkgs = import nixpkgs { inherit system; };\n"
	nixConfig += "        in pkgs.mkShell {\n"
	nixConfig += "          packages = with pkgs; [\n"

	for _, pkg := range cfg.Nix.Packages.Core {
		nixConfig += fmt.Sprintf("            %s\n", pkg)
	}
	if len(cfg.Nix.Packages.Optional) > 0 {
		for _, pkg := range cfg.Nix.Packages.Optional {
			nixConfig += fmt.Sprintf("            %s\n", pkg)
		}
	}

	nixConfig += "          ];\n"
	nixConfig += fmt.Sprintf("          name = \"%s-env\";\n", cfg.Metadata.Name)
	nixConfig += "        };\n"
	nixConfig += "      });\n"
	nixConfig += "    };\n"
	nixConfig += "}\n"

	return nixConfig
}

func (a *ApplyService) generateNixConfig(configPath string) (*schema.Config, string, error) {
	content, err := a.fs.ReadFile(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config: %w", err)
	}

	cfg, err := validator.ValidateYAMLContent(content)
	if err != nil {
		return nil, "", fmt.Errorf("invalid configuration: %w", err)
	}

	nixConfig := a.generateNixFromConfig(*cfg)
	return cfg, nixConfig, nil
}
