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
    ];

    userSettings = {
      "nix.enableLanguageServer" = true;
      "nix.serverPath" = "nil";
      "nix.serverSettings" = {
        "nil" = {
          "formatting" = { "command" = ["nixfmt"]; };
        };
      };
    };
  };
} 