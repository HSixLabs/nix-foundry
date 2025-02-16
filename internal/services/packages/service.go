package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
)

type Service interface {
	Add(packages []string, pkgType string) error
	Remove(packages []string, pkgType string) error
	List() (map[string][]string, error)
	Sync() error
}

type ServiceImpl struct {
	configDir string
	logger    *logging.Logger
}

func NewService(configDir string) Service {
	return &ServiceImpl{
		configDir: configDir,
		logger:    logging.GetLogger(),
	}
}

func (s *ServiceImpl) Add(packages []string, pkgType string) error {
	s.logger.Infof("Adding packages %v to type %s", packages, pkgType)

	// Ensure packages directory exists
	packagesDir := filepath.Join(s.configDir, "environments", "default", "packages")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Create or update the package type file
	pkgFile := filepath.Join(packagesDir, fmt.Sprintf("%s.nix", pkgType))

	// Read existing content if file exists
	var existingPkgs []string
	if content, err := os.ReadFile(pkgFile); err == nil {
		// Parse existing packages
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if pkg := strings.TrimSpace(line); pkg != "" && !strings.HasPrefix(pkg, "#") {
				existingPkgs = append(existingPkgs, pkg)
			}
		}
	}

	// Add new packages, avoiding duplicates
	pkgMap := make(map[string]bool)
	for _, pkg := range existingPkgs {
		pkgMap[pkg] = true
	}
	for _, pkg := range packages {
		pkgMap[pkg] = true
	}

	// Create the new file content
	var content strings.Builder
	content.WriteString("# This file is managed by nix-foundry\n")
	content.WriteString("{\n")
	content.WriteString("  environment.systemPackages = with pkgs; [\n")
	for pkg := range pkgMap {
		content.WriteString(fmt.Sprintf("    %s\n", pkg))
	}
	content.WriteString("  ];\n")
	content.WriteString("}\n")

	// Write the file
	if err := os.WriteFile(pkgFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write package file: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Remove(packages []string, pkgType string) error {
	s.logger.Infof("Removing packages %v from type %s", packages, pkgType)

	// Get package file path
	pkgFile := filepath.Join(s.configDir, "environments", "default", "packages", fmt.Sprintf("%s.nix", pkgType))

	// Read existing packages
	content, err := os.ReadFile(pkgFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("package file does not exist for type %s", pkgType)
		}
		return fmt.Errorf("failed to read package file: %w", err)
	}

	// Create set of packages to remove
	removeSet := make(map[string]bool)
	for _, pkg := range packages {
		removeSet[pkg] = true
	}

	// Filter out removed packages
	var remainingPkgs []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		pkg := strings.TrimSpace(line)
		if pkg != "" && !strings.HasPrefix(pkg, "#") && !strings.HasPrefix(pkg, "{") && !strings.HasPrefix(pkg, "}") {
			if !removeSet[pkg] {
				remainingPkgs = append(remainingPkgs, pkg)
			}
		}
	}

	// Create new file content
	var newContent strings.Builder
	newContent.WriteString("# This file is managed by nix-foundry\n")
	newContent.WriteString("{\n")
	newContent.WriteString("  environment.systemPackages = with pkgs; [\n")
	for _, pkg := range remainingPkgs {
		newContent.WriteString(fmt.Sprintf("    %s\n", pkg))
	}
	newContent.WriteString("  ];\n")
	newContent.WriteString("}\n")

	// Write updated file
	if err := os.WriteFile(pkgFile, []byte(newContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write package file: %w", err)
	}

	return nil
}

func (s *ServiceImpl) List() (map[string][]string, error) {
	s.logger.Info("Listing packages")

	result := make(map[string][]string)
	packagesDir := filepath.Join(s.configDir, "environments", "default", "packages")

	// Read package directory
	files, err := os.ReadDir(packagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("failed to read packages directory: %w", err)
	}

	// Process each .nix file
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".nix") {
			continue
		}

		pkgType := strings.TrimSuffix(file.Name(), ".nix")
		content, err := os.ReadFile(filepath.Join(packagesDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read package file %s: %w", file.Name(), err)
		}

		// Parse packages from file
		var packages []string
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "{") && !strings.HasPrefix(line, "}") {
				if pkg := strings.TrimSpace(strings.TrimPrefix(line, "environment.systemPackages")); pkg != "" {
					packages = append(packages, pkg)
				}
			}
		}

		result[pkgType] = packages
	}

	return result, nil
}

func (s *ServiceImpl) Sync() error {
	s.logger.Info("Synchronizing packages")

	// Ensure packages directory exists
	packagesDir := filepath.Join(s.configDir, "environments", "default", "packages")
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create packages directory: %w", err)
	}

	// Create default package files if they don't exist
	defaultTypes := []string{"core", "user", "team"}
	for _, pkgType := range defaultTypes {
		pkgFile := filepath.Join(packagesDir, fmt.Sprintf("%s.nix", pkgType))
		if _, err := os.Stat(pkgFile); os.IsNotExist(err) {
			// Create empty package file
			content := "# This file is managed by nix-foundry\n{\n  environment.systemPackages = with pkgs; [\n  ];\n}\n"
			if err := os.WriteFile(pkgFile, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create %s package file: %w", pkgType, err)
			}
		}
	}

	return nil
}
