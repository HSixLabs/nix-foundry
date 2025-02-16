package testutils

// MockFlakeContent is a minimal flake.nix content used for testing
const MockFlakeContent = `{
  description = "Test flake";
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
}`
