{ config, pkgs, lib, ... }:

{
  programs.git = {
    enable = true;
    userName = "Shawn Hoffman";
    userEmail = "your.email@example.com";
    
    extraConfig = {
      core = {
        editor = "nvim";
        autocrlf = "input";
      };
      init.defaultBranch = "main";
      pull.rebase = true;
      push.autoSetupRemote = true;
      
      url = {
        "ssh://git@github.com/" = {
          insteadOf = "https://github.com/";
        };
      };
    };
    
    ignores = [
      ".DS_Store"
      "*.swp"
      ".env"
      ".direnv"
      "node_modules"
    ];
  };
}
