package initservice_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	initservice "github.com/shawnkhoffman/nix-foundry/internal/services/init"
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
		projectSvc := project.NewService(
			config.NewService(),
			environment.NewService(
				tempDir,
				config.NewService(),
				platform.NewService(),
				true,
				true,
				true,
			),
			packages.NewService(tempDir),
		)

		// No need for type conversion since interfaces are aligned
		initService := initservice.NewService(tempDir, projectSvc)

		// First initialization
		if err := initService.Initialize(false); err != nil {
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
