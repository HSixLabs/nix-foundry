{ config, pkgs, lib, ... }:

let
  vscodeInstalled = builtins.pathExists "/Applications/Visual Studio Code.app" ||
                    builtins.pathExists "/usr/local/bin/code" ||
                    builtins.pathExists "/opt/homebrew/bin/code";
in
{
  programs.vscode = lib.mkIf vscodeInstalled {
    enable = true;
    package = null;  # Don't install VS Code through nix
    mutableExtensionsDir = true;
    enableUpdateCheck = false;
    enableExtensionUpdateCheck = false;
    
    extensions = with pkgs.vscode-extensions; [
      bbenoist.nix
      esbenp.prettier-vscode
      dbaeumer.vscode-eslint
      eamodio.gitlens
      
      # Languages & Frameworks
      ms-python.python
      ms-python.vscode-pylance
      golang.go
      
      # DevOps & Cloud
      ms-kubernetes-tools.vscode-kubernetes-tools
      ms-azuretools.vscode-docker
      hashicorp.terraform
      
      # Remote Development
      ms-vscode-remote.remote-ssh
      ms-vscode-remote.remote-containers
      ms-vscode-remote.remote-wsl
    ];

    userSettings = {
      "editor.fontFamily" = "'MesloLGS NF', 'Droid Sans Mono', 'monospace'";
      "editor.fontSize" = 14;
      "editor.formatOnSave" = true;
      "editor.renderWhitespace" = "all";
      "files.trimTrailingWhitespace" = true;
      "terminal.integrated.fontFamily" = "MesloLGS NF";
      "workbench.colorTheme" = "Default Dark+";
      "nix.enableLanguageServer" = true;
      "nix.serverPath" = "nil";
      "nix.serverSettings" = {
        "nil" = {
          "formatting" = { "command" = ["nixfmt"]; };
        };
      };
      "update.mode" = "none";
      "extensions.autoUpdate" = false;
    };
  };
} 