package config

import (
	"os"
	"path/filepath"
	"text/template"
)

const homeTemplate = `{ pkgs, ... }: {
  home.username = "{{.Username}}";
  home.homeDirectory = "{{.HomeDir}}";
  home.stateVersion = "23.11";

  # Let Home Manager manage itself
  programs.home-manager.enable = true;

  # Basic packages
  home.packages = with pkgs; [
    # Additional packages
    {{range .Packages.Additional}}
    {{.}}
    {{end}}

    # Development packages
    {{range .Packages.Development}}
    {{.}}
    {{end}}

    # Platform-specific packages
    {{if .Platform}}
    {{range index .Packages.PlatformSpecific .Platform.OS}}
    {{.}}
    {{end}}
    {{end}}

    # Team-specific packages
    {{if .Team.Enable}}
    {{range index .Packages.Team .Team.Name}}
    {{.}}
    {{end}}
    {{end}}
  ];

  {{if eq .Shell.Type "zsh"}}
  programs.zsh.enable = true;
  {{else if eq .Shell.Type "bash"}}
  programs.bash.enable = true;
  {{end}}

  {{if .Git.Enable}}
  programs.git = {
    enable = true;
    userName = "{{.Git.User.Name}}";
    userEmail = "{{.Git.User.Email}}";
    {{if .Git.Config}}
    extraConfig = {
      {{range $key, $value := .Git.Config}}
      {{$key}} = "{{$value}}";
      {{end}}
    };
    {{end}}
  };
  {{end}}

  # User customization through custom.nix
  imports = [
    (if builtins.pathExists ./custom.nix then ./custom.nix else {})
  ];
}`

type homeConfig struct {
	*NixConfig
	Username string
	HomeDir  string
}

func generateHome(configDir string, cfg *NixConfig) error {
	// Get user info
	username := os.Getenv("USER")
	if username == "" {
		username = "user"
	}
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home/" + username
	}

	homeCfg := &homeConfig{
		NixConfig: cfg,
		Username:  username,
		HomeDir:   homeDir,
	}

	f, err := os.Create(filepath.Join(configDir, "home.nix"))
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("home").Parse(homeTemplate))
	return tmpl.Execute(f, homeCfg)
}
