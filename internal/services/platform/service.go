package platform

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
)

type Service interface {
	SetupPlatform(testMode bool) error
	InstallHomeManager() error
	ValidateBackup(backupPath string) error
	RestoreFromBackup(backupPath, targetDir string) error
	EnableFlakeFeatures() error
}

type ServiceImpl struct {
	logger          *logging.Logger
	InstallHomebrew func() error
	os              string
}

func NewService() Service {
	return &ServiceImpl{
		logger: logging.GetLogger(),
		os:     runtime.GOOS,
	}
}

func (s *ServiceImpl) SetupPlatform(testMode bool) error {
	if testMode {
		s.logger.Debug("Skipping platform setup in test mode")
		return nil
	}

	if s.os == "darwin" {
		if err := s.setupDarwin(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServiceImpl) setupDarwin() error {
	// Ensure Homebrew is installed on macOS
	if _, err := exec.LookPath("brew"); err != nil {
		spin := progress.NewSpinner("Installing Homebrew...")
		spin.Start()
		if err := s.InstallHomebrew(); err != nil {
			spin.Fail("Failed to install Homebrew")
			return errors.NewPlatformError(err, "homebrew installation")
		}
		spin.Success("Homebrew installed")
	}
	return nil
}

func (s *ServiceImpl) InstallHomeManager() error {
	s.logger.Info("Installing home-manager")

	spin := progress.NewSpinner("Installing home-manager...")
	spin.Start()
	defer spin.Stop()

	// Install home-manager using nix-channel
	cmd := exec.Command("nix-channel", "--add", "https://github.com/nix-community/home-manager/archive/master.tar.gz", "home-manager")
	if err := cmd.Run(); err != nil {
		spin.Fail("Failed to install home-manager")
		return errors.NewPlatformError(err, "adding home-manager channel")
	}

	cmd = exec.Command("nix-channel", "--update")
	if err := cmd.Run(); err != nil {
		spin.Fail("Failed to update channels")
		return errors.NewPlatformError(err, "updating channels")
	}

	cmd = exec.Command("nix-shell", "<home-manager>", "-A", "installer", "--run", "home-manager init")
	if err := cmd.Run(); err != nil {
		spin.Fail("Failed to install home-manager")
		return errors.NewPlatformError(err, "initializing home-manager")
	}

	spin.Success("home-manager installed")
	return nil
}

func (s *ServiceImpl) ValidateBackup(backupPath string) error {
	// Implementation of ValidateBackup method
	return nil
}

func (s *ServiceImpl) RestoreFromBackup(backupPath, targetDir string) error {
	// Implementation of RestoreFromBackup method
	return os.Rename(backupPath, targetDir)
}

func (s *ServiceImpl) EnableFlakeFeatures() error {
	s.logger.Info("Enabling Nix flake features")

	// Create .config/nix directory if it doesn't exist
	nixConfigDir := os.Getenv("HOME") + "/.config/nix"
	if err := os.MkdirAll(nixConfigDir, 0755); err != nil {
		return errors.NewPlatformError(err, "failed to create nix config directory")
	}

	// Write nix.conf with flake features enabled
	nixConfPath := nixConfigDir + "/nix.conf"
	content := "experimental-features = nix-command flakes"
	if err := os.WriteFile(nixConfPath, []byte(content), 0644); err != nil {
		return errors.NewPlatformError(err, "failed to write nix.conf")
	}

	return nil
}
