{ config, pkgs, lib, ... }:

let
  inherit (lib) mkIf;
  isDarwin = pkgs.stdenv.isDarwin;
  vscodeInstalled = builtins.pathExists "/Applications/Visual Studio Code.app" ||
                    builtins.pathExists "/usr/local/bin/code" ||
                    builtins.pathExists "/opt/homebrew/bin/code";
                    
  dummyPackage = pkgs.runCommand "vscode-dummy" {} ''
    mkdir -p $out
    echo "1.0.0" > $out/version
  '';
  dummyDrv = dummyPackage // {
    pname = "vscode";
    version = "1.0.0";
  };
in
{
  programs.vscode = {
    enable = true;
    package = if isDarwin && vscodeInstalled
      then dummyDrv
      else if isDarwin
      then pkgs.vscode-darwin
      else pkgs.vscodium;
    
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
      "editor.fontFamily" = "'Hack Nerd Font', monospace";
      "editor.fontSize" = 14;
      "editor.lineNumbers" = "relative";
      "editor.renderWhitespace" = "boundary";
      "editor.rulers" = [ 80 120 ];
      "files.trimTrailingWhitespace" = true;
      "terminal.integrated.fontFamily" = "'Hack Nerd Font'";
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