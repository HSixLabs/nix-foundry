package packages

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
)

// New tests added for type-specific operations
func TestAddPackagesWithType(t *testing.T) {
	tempDir := t.TempDir()
	svc := packages.NewService(tempDir)

	t.Run("add single package", func(t *testing.T) {
		err := svc.Add([]string{"neovim"}, "core")
		if err != nil {
			t.Fatalf("Add() failed: %v", err)
		}

		// Verify package type and contents
		pkgs, err := svc.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if _, ok := pkgs["core"]; !ok {
			t.Error("Expected 'core' package type to be created")
		}

		found := false
		for _, pkg := range pkgs["core"] {
			if pkg == "neovim" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Package 'neovim' not found in core packages")
		}
	})
}
