{ config, pkgs, lib, ... }:

{
  programs.git = {
    enable = true;
    package = pkgs.git;
    
    extraConfig = {
      init.defaultBranch = "main";
      core.autocrlf = "input";
      pull.rebase = true;
    };
  };
}
