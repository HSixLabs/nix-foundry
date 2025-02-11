{ hostName, system ? "x86_64-linux" }:

let
  uname = let 
    user = builtins.getEnv "USER";
    username = builtins.getEnv "USERNAME";
    logname = builtins.getEnv "LOGNAME";
    nixUser = builtins.getEnv "NIX_USER";
  in
    if nixUser != "" then nixUser
    else if user != "" then user
    else if username != "" then username
    else if logname != "" then logname
    else throw ''
      No username found in environment variables.
      Please set NIX_USER environment variable before running home-manager.
      Example: export NIX_USER=$(whoami)
    '';
    
  isDarwin = system == "aarch64-darwin" || system == "x86_64-darwin";
  isWindows = system == "x86_64-windows";
  homePrefix = if isDarwin then "/Users"
              else if isWindows then "C:/Users"
              else "/home";
  
  _ = builtins.trace "Debug: NIX_USER=${builtins.getEnv "NIX_USER"}, USER=${builtins.getEnv "USER"}" null;
in
rec {
  username = uname;
  homeDirectory = if isWindows 
                 then "${homePrefix}\\${username}"
                 else "${homePrefix}/${username}";
  inherit hostName;
}