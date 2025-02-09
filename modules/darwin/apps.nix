{ config, pkgs, lib, ... }:

{
  # Homebrew configuration
  homebrew = {
    enable = true;
    onActivation = {
      autoUpdate = true;
      cleanup = "zap";
    };
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

      # Development
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

      # System Tools
      "bash-completion"
      "exiftool"
      "ffmpeg"
      "jq"
      "neofetch"
      "netpbm"
      "stow"
      "wireguard-tools"
      "wireshark"
      "yamllint"
    ];
    casks = [
      # Development Tools
      "android-studio"
      "cursor"
      "datagrip"
      "docker"
      "goland"
      "intellij-idea"
      "iterm2"
      "keybase"
      "miniconda"
      "powershell"
      "pycharm"
      "visual-studio-code"
      "webstorm"

      # Cloud & DevOps
      "google-cloud-sdk"
      "postman"
      "session-manager-plugin"

      # Browsers
      "microsoft-edge"

      # Fonts
      "font-awesome-terminal-fonts"
      "font-fontawesome"
      "font-hack-nerd-font"
      "font-meslo-lg-nerd-font"

      # Utilities
      "adobe-acrobat-pro"
      "anki"
      "charles"
      "chatgpt"
      "discord"
      "expressvpn"
      "gimp"
      "grammarly"
      "notion"
      "obs"
      "obsidian"
      "onedrive"
      "openvpn-connect"
      "signal"
      "slack"
      "spotify"
      "tunnelblick"
      "vlc"
      "wireshark"
    ];
  };
}
