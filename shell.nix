{ pkgs ? import <nixpkgs> { } }:

let
  nixpkgs-fmt-path = "${pkgs.nixpkgs-fmt}/bin/nixpkgs-fmt";
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    # Go development
    go
    gopls
    golangci-lint

    # Nix tools
    nixpkgs-fmt
    pre-commit

    # Testing tools
    go-tools # for testing utilities
    gotest
    gotestsum

    # Linting and formatting
    gofumpt
    golines

    # Build dependencies
    gcc
    git
  ];

  shellHook = ''
    echo "🛠️  nix-foundry Development Environment"
    echo "Available tools:"
    echo "  - Go $(go version | cut -d' ' -f3)"
    echo "  - golangci-lint $(golangci-lint --version | head -n1)"
    echo "  - nixpkgs-fmt (for Nix formatting)"
    echo "  - pre-commit (for git hooks)"
    echo ""
    echo "To set up git hooks, run: pre-commit install"

    # Set up environment variables
    export GOPATH="$PWD/.go"
    export PATH="$GOPATH/bin:$PATH"
    export NIXPKGS_FMT="${nixpkgs-fmt-path}"

    # Create local directories if they don't exist
    mkdir -p .go

    # Install pre-commit hooks if not already installed
    if [ ! -f .git/hooks/pre-commit ]; then
      pre-commit install
    fi
  '';
}
