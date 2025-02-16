package main

import (
	"os"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
