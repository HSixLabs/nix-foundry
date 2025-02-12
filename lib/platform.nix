{ lib }:

rec {
  isDarwin = system: builtins.match ".*-darwin" system != null;
  isLinux = system: builtins.match ".*-linux" system != null;
  isWindows = system: system == "x86_64-windows";
  isWSL = system: builtins.match ".*-linux-wsl" system != null;

  platformSettings = system: {
    homePrefix =
      if isDarwin system then "/Users"
      else if isWindows system then "C:/Users"
      else "/home";

    pathSeparator = if isWindows system then "\\" else "/";

    defaultShell =
      if isWindows system then "pwsh"
      else "zsh";

    configDir =
      if isWindows system then "$APPDATA/nix-foundry"
      else "$HOME/.config/nix-foundry";
  };
}
