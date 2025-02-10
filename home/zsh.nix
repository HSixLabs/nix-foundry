{ config, pkgs, lib, ... }:

{
  programs.zsh = {
    enable = true;
    autosuggestion.enable = true;
    enableCompletion = true;
    
    # Add history configuration
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
        src = pkgs.zsh-powerlevel10k;
        file = "share/zsh-powerlevel10k/powerlevel10k.zsh-theme";
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
          sha256 = "1h0vm2dgrmb8i2pvsgis3lshc5b0ad846836m62y8h3rdb3zmpy1";
        };
      }
    ];
  };

  # Use the p10k.nix configuration instead of external file
  xdg.configFile."zsh/.p10k.zsh".text = (import ./p10k.nix { inherit config pkgs lib; }).text;

  # Set ZDOTDIR explicitly
  home.sessionVariables = {
    ZDOTDIR = "$HOME/.config/zsh";
  };
} 