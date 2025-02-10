{ config, pkgs, lib, ... }:

{
  xdg.configFile."zsh/conf.d/vim-mode.zsh".text = ''
    # Load required module for menuselect keymap
    zmodload zsh/complist

    # Enable vi mode
    bindkey -v
    export KEYTIMEOUT=1

    # Use vim keys in tab complete menu
    bindkey -M menuselect 'h' vi-backward-char
    bindkey -M menuselect 'k' vi-up-line-or-history
    bindkey -M menuselect 'l' vi-forward-char
    bindkey -M menuselect 'j' vi-down-line-or-history

    # Change cursor shape for different vi modes
    function zle-keymap-select {
      if [[ ''${KEYMAP} == vicmd ]] ||
         [[ $1 = 'block' ]]; then
        echo -ne '\e[1 q'
      elif [[ ''${KEYMAP} == main ]] ||
           [[ ''${KEYMAP} == viins ]] ||
           [[ ''${KEYMAP} = "" ]] ||
           [[ $1 = 'beam' ]]; then
        echo -ne '\e[5 q'
      fi
    }
    zle -N zle-keymap-select
  '';
}