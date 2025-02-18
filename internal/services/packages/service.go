package packages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
)

// Service defines the package management operations
type Service interface {
	ListCustomPackages(ctx context.Context) ([]string, error)
	AddPackages(ctx context.Context, packages []string) error
	RemovePackages(ctx context.Context, packages []string) error
	Add(packages []string, pkgType string) error
	Remove(packages []string, pkgType string) error
	List() (map[string][]string, error)
	Sync() error
	Validate() error
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

func (s *ServiceImpl) ListCustomPackages(ctx context.Context) ([]string, error) {
	userPkgs, err := s.loadPackageType("user")
	if err != nil {
		return nil, fmt.Errorf("failed to load user packages: %w", err)
	}

	teamPkgs, err := s.loadPackageType("team")
	if err != nil {
		return nil, fmt.Errorf("failed to load team packages: %w", err)
	}

	return append(userPkgs, teamPkgs...), nil
}

func (s *ServiceImpl) AddPackages(ctx context.Context, packages []string) error {
	return s.modifyPackages("user", packages, true)
}

func (s *ServiceImpl) RemovePackages(ctx context.Context, packages []string) error {
	return s.modifyPackages("user", packages, false)
}

func (s *ServiceImpl) modifyPackages(pkgType string, packages []string, add bool) error {
	existing, err := s.loadPackageType(pkgType)
	if err != nil {
		return err
	}

	packageMap := make(map[string]bool)
	for _, pkg := range existing {
		packageMap[pkg] = true
	}

	for _, pkg := range packages {
		if add {
			packageMap[pkg] = true
		} else {
			delete(packageMap, pkg)
		}
	}

	var final []string
	for pkg := range packageMap {
		final = append(final, pkg)
	}

	return s.savePackageType(pkgType, final)
}

func (s *ServiceImpl) loadPackageType(pkgType string) ([]string, error) {
	filePath := filepath.Join(s.configDir, "environments", "default", "packages", pkgType+".nix")

	content, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}

	return parseNixPackages(string(content)), nil
}

func parseNixPackages(content string) []string {
	var packages []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "pkgs.") {
			pkg := strings.TrimPrefix(trimmed, "pkgs.")
			pkg = strings.TrimRight(pkg, ";")
			packages = append(packages, pkg)
		}
	}
	return packages
}

func (s *ServiceImpl) savePackageType(pkgType string, packages []string) error {
	filePath := filepath.Join(s.configDir, "environments", "default", "packages", pkgType+".nix")

	var content strings.Builder
	content.WriteString("# Managed by nix-foundry - DO NOT EDIT DIRECTLY\n")
	content.WriteString("{ config, pkgs, ... }: {\n")
	content.WriteString("  environment.systemPackages = with pkgs; [\n")

	for _, pkg := range packages {
		content.WriteString(fmt.Sprintf("    %s\n", pkg))
	}

	content.WriteString("  ];\n}\n")

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

func (s *ServiceImpl) Validate() error {
	// Implementation of Validate method
	return nil
}
