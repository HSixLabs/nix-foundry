{ config, pkgs, lib, ... }:

{
  homebrew = {
    enable = true;
    onActivation = {
      autoUpdate = true;
      cleanup = "zap";
      upgrade = true;
    };
    
    # Generate Brewfile on changes
    global = {
      brewfile = true;
      autoUpdate = true;
      lockfiles = {
        enable = true;
        path = "~/.config/Brewfile";
      };
    };

    # Merge taps from both files
    taps = [
      "conduktor/brew"
      "derailed/k9s"
      "hashicorp/tap"
      "homebrew/autoupdate"
      "homebrew/bundle"
      "homebrew/cask-fonts"
      "homebrew/services"
      "norwoodj/tap"
    ];

    # Merge brews from both files
    brews = [
      # Cloud & Infrastructure
      "argocd"
      "awscli"
      "azure-cli"
      "certbot"
      "cfssl"
      "chart-testing"
      "docker-compose"
      "doctl"
      "docutils"
      "eksctl"
      "helm"
      "kubeconform"
      "kubernetes-cli"
      "minikube"
      "pulumi"
      "tfenv"
      "derailed/k9s/k9s"
      "norwoodj/tap/helm-docs"

      # Development tools from apps.nix
      "cairo"
      "git"
      "go"
      "harfbuzz"
      "icu4c@75"
      "node"
      "openjdk"
      "pipx"
      "progress"
      "pyenv"
      "python@3.10"
      "python@3.11"
      "yarn"
    ];
  };

  # Update activation script to handle Brewfile properly
  system.activationScripts.postActivation.text = lib.mkAfter ''
    # Backup existing Brewfile
    if [ -f ~/.config/Brewfile ]; then
      cp ~/.config/Brewfile ~/.config/Brewfile.backup
    fi
    
    # Install packages from Brewfile if it exists
    if [ -f ~/.config/Brewfile ]; then
      echo "Installing Homebrew packages from Brewfile..."
      /opt/homebrew/bin/brew bundle install --file=~/.config/Brewfile || true
    fi
  '';
} 