/*
Package main provides documentation generation functionality.
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/docs"
)

func main() {
	if err := generateDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
		os.Exit(1)
	}
}

/*
generateDocs generates all documentation files for Nix Foundry.
*/
func generateDocs() error {
	if err := os.MkdirAll("docs", 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	guides := map[string]func() (string, error){
		"installation.md":    docs.GenerateInstallGuide,
		"packages.md":        docs.GeneratePackageGuide,
		"troubleshooting.md": docs.GenerateTroubleshootingGuide,
		"uninstall.md":       docs.GenerateUninstallGuide,
		"platform-support.md": func() (string, error) {
			return docs.GeneratePlatformList(), nil
		},
	}

	for filename, genFunc := range guides {
		content, err := genFunc()
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}

		content = strings.TrimSpace(content)
		data := make([]byte, 0, len(content)+1)
		data = append(data, content...)
		data = append(data, '\n')

		path := filepath.Join("docs", filename)
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		fmt.Printf("✨ Generated %s\n", path)
	}

	return nil
}
