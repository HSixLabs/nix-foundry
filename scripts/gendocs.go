/*
Package main provides documentation generation functionality.
*/
package main

import (
	"fmt"
	"os"

	"github.com/shawnkhoffman/nix-foundry/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	if err := generateDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
		os.Exit(1)
	}
}

/*
generateDocs generates documentation for all CLI commands.
It uses Cobra's built-in doc generation to create markdown files
for each command in the CLI.
*/
func generateDocs() error {
	if err := os.MkdirAll("docs", 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	root := cmd.GetRootCommand()
	if err := doc.GenMarkdownTree(root, "docs"); err != nil {
		return fmt.Errorf("failed to generate command documentation: %w", err)
	}

	return nil
}
