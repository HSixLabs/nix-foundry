package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/nix"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system diagnostics",
	Long:  `Run diagnostics to check system health and configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiagnostics()
	},
}

func runDiagnostics() error {
	fmt.Println("🔍 Running system diagnostics...")

	// 1. Check system requirements
	if err := checkSystemRequirements(); err != nil {
		fmt.Printf("❌ System requirements: %v\n", err)
		if fix {
			if err := fixSystemRequirements(); err != nil {
				return fmt.Errorf("failed to fix system requirements: %w", err)
			}
			fmt.Println("✅ Fixed system requirements")
		}
	} else {
		fmt.Println("✅ System requirements: OK")
	}

	// 2. Check configuration
	if configManager == nil {
		var err error
		configManager, err = config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}
	}

	var nixConfig config.NixConfig
	if err := configManager.ReadConfig("config.yaml", &nixConfig); err != nil {
		fmt.Printf("❌ Configuration: %v\n", err)
		return err
	}

	validator := config.NewValidator(&nixConfig)
	if err := validator.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}
	fmt.Println("✅ Configuration: OK")

	// 3. Platform-specific checks
	if err := runPlatformChecks(); err != nil {
		return err
	}

	// 4. Security audit if requested
	if security {
		if err := runSecurityAudit(); err != nil {
			return err
		}
	}

	fmt.Println("✅ Diagnostics completed successfully")
	return nil
}

func fixSystemRequirements() error {
	// Install missing dependencies
	if _, err := exec.LookPath("nix"); err != nil {
		if err := installNix(); err != nil {
			return fmt.Errorf("failed to install nix: %w", err)
		}
	}

	if _, err := exec.LookPath("home-manager"); err != nil {
		if err := installHomeManager(); err != nil {
			return fmt.Errorf("failed to install home-manager: %w", err)
		}
	}

	return nil
}

func runPlatformChecks() error {
	// Basic platform checks
	if runtime.GOOS == "darwin" {
		if err := checkHomebrew(); err != nil {
			fmt.Printf("⚠️  Homebrew check: %v\n", err)
		} else {
			fmt.Println("✅ Homebrew: OK")
		}
	}

	if runtime.GOOS == "linux" {
		if err := checkSELinux(); err != nil {
			fmt.Printf("⚠️  SELinux check: %v\n", err)
		}
		if err := checkWSL(); err != nil {
			fmt.Printf("⚠️  WSL check: %v\n", err)
		}
	}

	return nil
}

func checkHomebrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not installed")
	}
	return nil
}

func checkSELinux() error {
	if _, err := os.Stat("/etc/selinux/config"); err != nil {
		return fmt.Errorf("SELinux not configured")
	}
	return nil
}

func checkWSL() error {
	if _, err := os.ReadFile("/proc/version"); err != nil {
		return fmt.Errorf("not running in WSL")
	}
	return nil
}

func checkPermissions() error {
	configDir := getConfigDir()

	// Check directory permissions
	info, err := os.Stat(configDir)
	if err != nil {
		return fmt.Errorf("failed to check config directory: %w", err)
	}

	// Ensure directory is not world-writable
	if info.Mode().Perm()&0002 != 0 {
		return fmt.Errorf("config directory has unsafe permissions")
	}

	return nil
}

func runSecurityAudit() error {
	fmt.Println("\n🔒 Running security audit...")

	// Check configuration directory permissions
	if err := checkPermissions(); err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	fmt.Println("✅ Configuration permissions: OK")

	// Check environment isolation
	if err := checkEnvironmentIsolation(); err != nil {
		return fmt.Errorf("environment isolation check failed: %w", err)
	}
	fmt.Println("✅ Environment isolation: OK")

	// Check backup integrity
	configDir := getConfigDir()
	backupDir := filepath.Join(configDir, "backups")
	if _, err := os.Stat(backupDir); err == nil {
		if err := checkBackupPermissions(backupDir); err != nil {
			return fmt.Errorf("backup security check failed: %w", err)
		}
		fmt.Println("✅ Backup security: OK")
	}

	fmt.Println("✅ Security audit completed successfully")
	return nil
}

func installNix() error {
	fmt.Println("Installing Nix...")

	// Use the nix package's Install function
	if err := nix.Install(); err != nil {
		return fmt.Errorf("nix installation failed: %w", err)
	}

	// Verify installation
	if _, err := exec.LookPath("nix"); err != nil {
		return fmt.Errorf("nix installation verification failed")
	}

	return nil
}

func checkBackupPermissions(backupDir string) error {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get backup info: %w", err)
		}

		// Ensure backups are not world-readable or writable
		if info.Mode().Perm()&0077 != 0 {
			return fmt.Errorf("backup %s has unsafe permissions", entry.Name())
		}
	}

	return nil
}
