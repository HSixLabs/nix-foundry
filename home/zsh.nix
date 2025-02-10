{ config, pkgs, lib, ... }:

{
  imports = [ ./p10k.nix ];

  programs.zsh = {
    enable = true;
    autosuggestion.enable = true;
    enableCompletion = true;
    
    history = {
      size = 1000000;
      save = 1000000;
      path = "$HOME/.zsh_history";
      ignoreDups = true;
      share = true;
      extended = true;
    };

    initExtra = ''
      # Enable Powerlevel10k instant prompt
      if [[ -r "''${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-''${(%):-%n}.zsh" ]]; then
        source "''${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-''${(%):-%n}.zsh"
      fi

      # Source p10k config
      [[ ! -f $ZDOTDIR/.p10k.zsh ]] || source $ZDOTDIR/.p10k.zsh

      # ZSH options from your previous config
      setopt autocd extendedglob nomatch menucomplete
      setopt interactive_comments
      unsetopt BEEP

      # Useful ZLE configurations
      zle_highlight=('paste:none')
      
      # Key bindings for history search
      autoload -U up-line-or-beginning-search
      autoload -U down-line-or-beginning-search
      zle -N up-line-or-beginning-search
      zle -N down-line-or-beginning-search

      # Load colors
      autoload -Uz colors && colors

      # Completion styling
      zstyle ':completion:*' menu select
      _comp_options+=(globdots) # Include hidden files in completion
    '';
    
    plugins = [
      {
        name = "powerlevel10k";
        src = pkgs.fetchFromGitHub {
          owner = "romkatv";
          repo = "powerlevel10k";
          rev = "35833ea15f14b71dbcebc7e54c104d8d56ca5268";
          sha256 = "16rkmnak279xwi2qb3h2rk2940czg193mhim25lf61jvd8nn1k4a";
        };
        file = "powerlevel10k.zsh-theme";
      }
      {
        name = "zsh-autosuggestions";
        src = pkgs.zsh-autosuggestions;
        file = "share/zsh-autosuggestions/zsh-autosuggestions.zsh";
      }
      {
        name = "zsh-syntax-highlighting";
        src = pkgs.zsh-syntax-highlighting;
        file = "share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh";
      }
      {
        name = "zsh-autopair";
        src = pkgs.fetchFromGitHub {
          owner = "hlissner";
          repo = "zsh-autopair";
          rev = "34a8bca0c18fcf3ab1561caef9790abffc1d3d49";
          sha256 = "1x16y24hbwcaxfhqabw4x26jmpxzz2zzmlvs9nnbzaxyi20cwfyz";
        };
      }
    ];

    dotDir = ".config/zsh";
  };

  home.sessionVariables = {
    ZDOTDIR = "$HOME/.config/zsh";
  };
} 