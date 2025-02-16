package update

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
)

type Service interface {
	UpdateFlake(configDir string) error
	ApplyConfiguration(configDir string, testMode bool) error
}

type ServiceImpl struct {
	logger *logging.Logger
}

func NewService() Service {
	return &ServiceImpl{
		logger: logging.GetLogger(),
	}
}

func (s *ServiceImpl) UpdateFlake(configDir string) error {
	s.logger.Info("Updating Nix flake")

	spin := progress.NewSpinner("Updating Nix packages...")
	spin.Start()
	defer spin.Stop()

	// Check if configuration exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		spin.Fail("Failed to find configuration")
		return errors.NewValidationError(configDir, err,
			"no configuration found. Please run 'nix-foundry init' first")
	}

	// Convert the path to an absolute path
	absPath, err := filepath.Abs(configDir)
	if err != nil {
		spin.Fail("Failed to resolve config path")
		return errors.NewLoadError(configDir, err, "failed to resolve config path")
	}

	// Update flake
	updateCmd := exec.Command("nix", "flake", "update", "--flake", absPath)
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		spin.Fail("Failed to update packages")
		return errors.NewLoadError(absPath, err, "failed to update flake")
	}

	spin.Success("Packages updated")
	return nil
}

func (s *ServiceImpl) ApplyConfiguration(configDir string, testMode bool) error {
	s.logger.Info("Applying updated configuration")

	// Check if configuration exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return errors.NewValidationError(configDir, err,
			"no configuration found. Please run 'nix-foundry init' first")
	}

	spin := progress.NewSpinner("Applying configuration...")
	spin.Start()
	defer spin.Stop()

	// Convert the path to an absolute path
	absPath, err := filepath.Abs(configDir)
	if err != nil {
		return errors.NewLoadError(configDir, err, "failed to resolve config path")
	}

	// Apply configuration using home-manager
	cmd := exec.Command("home-manager", "switch", "--flake", absPath)
	if testMode {
		s.logger.Debug("Skipping configuration apply in test mode")
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		spin.Fail("Failed to apply configuration")
		return errors.NewLoadError(absPath, err, "failed to apply configuration")
	}

	spin.Success("Configuration applied successfully")
	s.logger.Info("✨ Environment updated successfully!")
	return nil
}
