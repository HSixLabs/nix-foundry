package update

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
)

// Service defines the interface for update operations
type Service interface {
	UpdateFlake(dir string) error
	ApplyConfiguration(dir string, dryRun bool) error
}

// ServiceImpl implements the update Service interface
type ServiceImpl struct {
	configService config.Service
	envService    environment.Service
	logger        *logging.Logger
}

// NewService creates a new update service instance
func NewService(configService config.Service, envService environment.Service) Service {
	return &ServiceImpl{
		configService: configService,
		envService:    envService,
		logger:        logging.GetLogger(),
	}
}

// UpdateFlake updates the Nix flake in the specified directory
func (s *ServiceImpl) UpdateFlake(dir string) error {
	s.logger.Info("Updating Nix flake")

	// Verify flake.nix exists
	flakePath := filepath.Join(dir, "flake.nix")
	if _, err := os.Stat(flakePath); err != nil {
		return errors.NewValidationError("update", err, "flake.nix not found")
	}

	// Run nix flake update
	cmd := exec.Command("nix", "flake", "update", dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.NewOperationError("update", err, map[string]interface{}{
			"output": string(output),
			"dir":    dir,
		})
	}

	s.logger.Debug("Flake update completed successfully")
	return nil
}

// ApplyConfiguration applies the Nix configuration using home-manager
func (s *ServiceImpl) ApplyConfiguration(dir string, dryRun bool) error {
	s.logger.Info("Applying Nix configuration")

	// Verify flake.nix exists
	flakePath := filepath.Join(dir, "flake.nix")
	if _, err := os.Stat(flakePath); err != nil {
		return errors.NewValidationError("update", err, "flake.nix not found")
	}

	// Build command arguments
	args := []string{"switch"}
	if dryRun {
		args = append(args, "--dry-run")
	}
	args = append(args, "--flake", dir)

	// Run home-manager switch
	cmd := exec.Command("home-manager", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.NewOperationError("update", err, map[string]interface{}{
			"output": string(output),
			"dir":    dir,
			"dryRun": dryRun,
		})
	}

	s.logger.Debug("Configuration applied successfully")
	return nil
}
