{ config, pkgs, lib, ... }:

{
  xdg.configFile."zsh/zsh-aliases".text = ''
    # Navigation
    alias ..='cd ..'
    alias ...='cd ../..'
    alias .3='cd ../../..'
    alias .4='cd ../../../..'

    # Shortcuts
    alias c='clear'
    alias e='exit'
    alias k='kubectl'
    alias g='git'
    alias vim='nvim'
    alias v='nvim'

    # ls aliases
    alias ls='ls --color=auto'
    alias ll='ls -lah'
    alias la='ls -A'
    alias l='ls -CF'

    # Git aliases
    alias gs='git status'
    alias ga='git add'
    alias gc='git commit'
    alias gp='git push'
    alias gl='git pull'
  '';
} 