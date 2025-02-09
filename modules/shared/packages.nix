{ config, pkgs, ... }:

{
  # Core packages shared between all systems
  environment.systemPackages = with pkgs; [
    # Core utilities
    coreutils
    curl
    wget
    vim
    neovim
    
    # Development
    git
    git-lfs
    gh
    ripgrep
    fd
    jq
    tree
    
    # Cloud & Infrastructure
    awscli2
    kubectl
    kubernetes-helm
    docker
    docker-compose
    terraform
    
    # Languages & Runtimes
    go
    nodejs
    python311
    
    # System utilities
    htop
    neofetch
    bash-completion
    yamllint
  ];

  # Allow unfree packages globally
  nixpkgs.config.allowUnfree = true;
} 