{ pkgs ? import <nixpkgs> { } }:

let
  nixpkgs-fmt-path = "${pkgs.nixpkgs-fmt}/bin/nixpkgs-fmt";
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    pre-commit
    nixpkgs-fmt
  ];

  shellHook = ''
    echo "Development environment loaded"
    echo "Run 'pre-commit install' to set up git hooks"
    export NIXPKGS_FMT="${nixpkgs-fmt-path}"
  '';
}
