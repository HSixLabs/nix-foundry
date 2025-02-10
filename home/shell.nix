{ config, pkgs, lib, ... }:

{
  programs.zsh = {
    enable = true;
    syntaxHighlighting.enable = true;
    autosuggestion.enable = true;
    
    plugins = [
      {
        name = "powerlevel10k";
        src = pkgs.fetchFromGitHub {
          owner = "romkatv";
          repo = "powerlevel10k";
          rev = "v1.20.0";
          sha256 = "0m1j1npx4vx8cl72w8jh7d52gxnhzp8w2w5l1lj6x0g5g9j4prb4";
        };
        file = "powerlevel10k.zsh-theme";
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
    
    initExtra = ''
      # Enable Powerlevel10k instant prompt
      if [[ -r "''${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-''${(%):-%n}.zsh" ]]; then
        source "''${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-''${(%):-%n}.zsh"
      fi

      # Source p10k config
      [[ ! -f $ZDOTDIR/.p10k.zsh ]] || source $ZDOTDIR/.p10k.zsh
    '';
  };

  # Add a basic p10k configuration directly in the nix configuration
  home.file.".p10k.zsh".text = ''
    # Generated p10k configuration
    'builtin' 'local' '-a' 'p10k_config_opts'
    [[ ! -o 'aliases'         ]] || p10k_config_opts+=('aliases')
    [[ ! -o 'sh_glob'        ]] || p10k_config_opts+=('sh_glob')
    [[ ! -o 'no_brace_expand' ]] || p10k_config_opts+=('no_brace_expand')
    'builtin' 'setopt' 'no_aliases' 'no_sh_glob' 'brace_expand'

    () {
      emulate -L zsh -o extended_glob
      unset -m '(POWERLEVEL9K_*|DEFAULT_USER)~POWERLEVEL9K_GITSTATUS_DIR'
      
      # Basic prompt configuration
      POWERLEVEL9K_MODE=nerdfont-complete
      POWERLEVEL9K_PROMPT_ON_NEWLINE=true
      POWERLEVEL9K_MULTILINE_FIRST_PROMPT_PREFIX=""
      POWERLEVEL9K_MULTILINE_LAST_PROMPT_PREFIX="%F{blue}❯%f "
      
      # Left prompt segments
      POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(
        dir                     # current directory
        vcs                     # git status
        command_execution_time  # previous command duration
        virtualenv             # python virtual environment
        status                 # exit code of the last command
      )
      
      # Right prompt segments
      POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(
        time                   # current time
      )
      
      # Basic styling
      POWERLEVEL9K_SHORTEN_DIR_LENGTH=2
      POWERLEVEL9K_SHORTEN_STRATEGY="truncate_middle"
    }
  '';
}
