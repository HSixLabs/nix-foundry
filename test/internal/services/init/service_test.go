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

func TestInitializeEnvironment(t *testing.T) {
	t.Run("fresh initialization", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := initservice.NewService(
			tempDir,
			project.NewService(
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
			),
		)

		// Helper functions for test assertions
		assertDirExists := func(path string) {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Directory %s does not exist", path)
			}
		}

		assertSymlinkTarget := func(link, target string) {
			actual, err := os.Readlink(link)
			if err != nil {
				t.Errorf("Failed to read symlink: %v", err)
			}
			if actual != target {
				t.Errorf("Symlink target mismatch: expected %s, got %s", target, actual)
			}
		}

		err := svc.Initialize(true)
		if err != nil {
			t.Fatalf("Initialization failed: %v", err)
		}

		assertDirExists(filepath.Join(tempDir, "environments"))
		assertDirExists(filepath.Join(tempDir, "backups"))
		assertSymlinkTarget(
			filepath.Join(tempDir, "environments", "current"),
			filepath.Join(tempDir, "environments", "default"),
		)
	})

	t.Run("force reinitialization", func(t *testing.T) {
		tempDir := t.TempDir()
		svc := initservice.NewService(
			tempDir,
			project.NewService(
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
			),
		)

		// First initialization
		if err := svc.Initialize(false); err != nil {
			t.Fatal(err)
		}

		// Force re-init
		err := svc.Initialize(true)
		if err != nil {
			t.Fatalf("Forced reinitialization failed: %v", err)
		}
	})
}
