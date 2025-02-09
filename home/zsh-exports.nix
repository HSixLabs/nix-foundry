{ config, pkgs, lib, ... }:

{
  xdg.configFile."zsh/zsh-exports".text = ''
    # XDG paths
    export XDG_CONFIG_HOME="$HOME/.config"
    export XDG_DATA_HOME="$HOME/.local/share"
    export XDG_CACHE_HOME="$HOME/.cache"

    # Path
    export PATH="$HOME/.local/bin:$PATH"
    
    # Editor
    export EDITOR="nvim"
    export VISUAL="code"
    export TERMINAL="alacritty"
    export BROWSER="brave"
    export MANPAGER='nvim +Man!'
    export MANWIDTH=999
    export PATH=$HOME/.cargo/bin:$PATH
    export PATH=$HOME/.local/share/go/bin:$PATH
    export GOPATH=$HOME/.local/share/go
    export PATH=$HOME/.fnm:$PATH
    export XDG_CURRENT_DESKTOP="Wayland"
  '';
} 