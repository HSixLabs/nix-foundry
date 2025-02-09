{ hostName }:

let
  uname = builtins.getEnv "USER";
  defaultUser = builtins.getEnv "NIX_DEFAULT_USER";
  # Add debug trace
  _ = builtins.trace "USER env var: ${builtins.toString uname}" null;
in
rec {
  username = if uname != "" then uname 
             else if defaultUser != "" then defaultUser
             else "shoffman";
  homeDirectory = "/Users/${username}";
  inherit hostName;
} 