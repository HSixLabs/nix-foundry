package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

func TestInitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("full initialization flow", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := NewService(
			tempDir,
			project.NewService(
				config.NewService(),
				environment.NewService(tempDir, config.NewService(), validation.NewService(), platform.NewService()),
				packages.NewService(tempDir),
			),
		)

		// First initialization
		if err := svc.Initialize(false); err != nil {
			t.Fatal(err)
		}

		// Verify critical files
		requiredFiles := []string{
			"config.yaml",
			"environments/default/flake.nix",
			"environments/default/packages/base.nix",
		}

		for _, file := range requiredFiles {
			fp := filepath.Join(tempDir, file)
			if _, err := os.Stat(fp); os.IsNotExist(err) {
				t.Errorf("Missing required file: %s", fp)
			}
		}
	})
}
