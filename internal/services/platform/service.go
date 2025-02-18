package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	IsHomeManagerInstalled() bool
	InstallNix() error
	IsNixInstalled() bool
	Validate() error
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
	s.logger.Debug("Checking Homebrew installation")

	if _, err := exec.LookPath("brew"); err != nil {
		spin := progress.NewSpinner("Installing Homebrew...")
		spin.Start()
		if err := s.InstallHomebrew(); err != nil {
			spin.Fail("Failed to install Homebrew")
			return errors.NewPlatformError(err, "homebrew installation")
		}
		spin.Success("Homebrew installed")
		s.logger.Debug("Homebrew installation completed")
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

func (s *ServiceImpl) IsHomeManagerInstalled() bool {
	s.logger.Debug("Checking if home-manager is installed")

	// Check if home-manager binary exists in PATH
	if _, err := exec.LookPath("home-manager"); err == nil {
		return true
	}

	// Check if home-manager channel is added
	cmd := exec.Command("nix-channel", "--list")
	output, err := cmd.Output()
	if err != nil {
		s.logger.Debug("Failed to check nix channels", "error", err)
		return false
	}

	return strings.Contains(string(output), "home-manager")
}

func (s *ServiceImpl) InstallNix() error {
	s.logger.Debug("Starting Nix installation")

	installCmd := exec.Command("sh", "-c", "curl -L https://nixos.org/nix/install | sh -s -- --no-daemon")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("installation script failed: %w", err)
	}

	// Update PATH for current process
	os.Setenv("PATH", fmt.Sprintf("%s:%s",
		"/nix/var/nix/profiles/default/bin",
		filepath.Join(os.Getenv("HOME"), ".nix-profile/bin"),
	) + ":" + os.Getenv("PATH"))

	return nil
}

func (s *ServiceImpl) updateShellEnvironment() error {
	// Update PATH for current process
	home := os.Getenv("HOME")
	newPath := fmt.Sprintf("%s/.nix-profile/bin:%s", home, os.Getenv("PATH"))
	if err := os.Setenv("PATH", newPath); err != nil {
		return err
	}

	// Update shell profile files
	profileFiles := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	envLine := `if [ -e $HOME/.nix-profile/etc/profile.d/nix.sh ]; then . $HOME/.nix-profile/etc/profile.d/nix.sh; fi`

	for _, file := range profileFiles {
		if _, err := os.Stat(file); err == nil {
			f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				continue
			}
			fmt.Fprintf(f, "\n%s\n", envLine)
			f.Close()
		}
	}

	return nil
}

func (s *ServiceImpl) IsNixInstalled() bool {
	s.logger.Debug("Checking Nix installation")
	// Check standard Nix locations
	nixPaths := []string{
		"/nix/var/nix/profiles/default/bin/nix",
		filepath.Join(os.Getenv("HOME"), ".nix-profile/bin/nix"),
		"/run/current-system/sw/bin/nix", // For NixOS systems
	}

	for _, path := range nixPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Fallback to checking if nix command works
	cmd := exec.Command("nix", "--version")
	return cmd.Run() == nil
}

func (s *ServiceImpl) Validate() error {
	// Add platform-specific validation logic
	return nil
}
