{ config, pkgs, lib, ... }:

{
  xdg.configFile."zsh/zsh-prompt".text = ''
    # This is handled by powerlevel10k now
    autoload -Uz promptinit
    promptinit
  '';
} 