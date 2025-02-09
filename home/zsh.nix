{ config, pkgs, lib, users, ... }:

{
  programs.zsh = {
    enable = true;
    autosuggestion.enable = true;
    enableCompletion = true;
    
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

    initExtra = ''
      # Enable Powerlevel10k instant prompt
      if [[ -r "''${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-''${(%):-%n}.zsh" ]]; then
        source "''${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-''${(%):-%n}.zsh"
      fi

      # Source user's custom configs
      for conf in $HOME/.config/zsh/conf.d/*.zsh(N); do
        source $conf
      done

      # Source local overrides if they exist
      if [[ -f $HOME/.zshrc.local ]]; then
        source $HOME/.zshrc.local
      fi
    '';

    dotDir = ".config/zsh";
  };
} 