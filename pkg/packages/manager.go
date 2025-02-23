/*
Package packages provides functionality for managing Nix packages across different platforms,

handling platform-specific package names, validations, and package groupings.
*/
package packages

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/platform"
)

/*
Manager handles platform-specific package management operations.
It provides functionality for package name resolution, validation,
and default package selections based on the current platform.
*/
type Manager struct {
	platform platform.Platform
	isWSL    bool
	fs       filesystem.FileSystem
}

// NewManager creates a new package manager instance.
func NewManager(fs filesystem.FileSystem) *Manager {
	return &Manager{
		platform: platform.GetPlatform(),
		isWSL:    platform.IsWSL(),
		fs:       fs,
	}
}

// InstallPackage installs a package using nix-env.
func (m *Manager) InstallPackage(pkg string) error {
	platformPkg := m.GetPackageName(pkg)
	cmd := exec.Command("nix-env", "-iA", "nixpkgs."+platformPkg)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install package %s: %s (%w)", pkg, string(output), err)
	}
	return nil
}

// RemovePackage removes a package using nix-env.
func (m *Manager) RemovePackage(pkg string) error {
	platformPkg := m.GetPackageName(pkg)
	cmd := exec.Command("nix-env", "-e", platformPkg)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove package %s: %s (%w)", pkg, string(output), err)
	}
	return nil
}

// ListInstalledPackages returns a list of installed packages.
func (m *Manager) ListInstalledPackages() ([]string, error) {
	cmd := exec.Command("nix-env", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var packages []string
	for _, pkg := range strings.Split(string(output), "\n") {
		if pkg = strings.TrimSpace(pkg); pkg != "" {
			packages = append(packages, pkg)
		}
	}
	return packages, nil
}

// SearchPackages searches for available packages matching the query.
func (m *Manager) SearchPackages(query string) (map[string]string, error) {
	cmd := exec.Command("nix-env", "-qaP", "--description", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	results := make(map[string]string)
	for _, line := range strings.Split(string(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 2 {
				name := strings.TrimPrefix(parts[1], "nixpkgs.")
				desc := ""
				if len(parts) > 2 {
					desc = strings.TrimSpace(parts[2])
				}
				results[name] = desc
			}
		}
	}
	return results, nil
}

// GetPackageName returns the platform-specific package name.
func (m *Manager) GetPackageName(pkg string) string {
	platformSpecificNames := map[platform.Platform]map[string]string{
		platform.MacOS: {
			"python":    "python3",
			"gcc":       "gcc_12",
			"vim":       "vim-darwin-huge",
			"git":       "git-darwin",
			"tmux":      "tmux-darwin",
			"openssh":   "openssh-darwin",
			"bash":      "bash-darwin",
			"zsh":       "zsh-darwin",
			"fish":      "fish-darwin",
			"curl":      "curl-darwin",
			"wget":      "wget-darwin",
			"rsync":     "rsync-darwin",
			"htop":      "htop-darwin",
			"neovim":    "neovim-darwin",
			"emacs":     "emacs-darwin",
			"nano":      "nano-darwin",
			"less":      "less-darwin",
			"grep":      "grep-darwin",
			"sed":       "gnused",
			"awk":       "gawk",
			"coreutils": "coreutils-darwin",
		},
		platform.Linux: {
			"python":    "python3",
			"gcc":       "gcc",
			"vim":       "vim",
			"git":       "git",
			"tmux":      "tmux",
			"openssh":   "openssh",
			"bash":      "bash",
			"zsh":       "zsh",
			"fish":      "fish",
			"curl":      "curl",
			"wget":      "wget",
			"rsync":     "rsync",
			"htop":      "htop",
			"neovim":    "neovim",
			"emacs":     "emacs",
			"nano":      "nano",
			"less":      "less",
			"grep":      "grep",
			"sed":       "sed",
			"awk":       "gawk",
			"coreutils": "coreutils",
		},
	}

	if m.isWSL {
		if linuxNames, ok := platformSpecificNames[platform.Linux]; ok {
			if name, ok := linuxNames[pkg]; ok {
				return name
			}
		}
		return pkg
	}

	if platformNames, ok := platformSpecificNames[m.platform]; ok {
		if name, ok := platformNames[pkg]; ok {
			return name
		}
	}

	return pkg
}

// GetDefaultPackages returns the default packages for the current platform.
func (m *Manager) GetDefaultPackages() []string {
	commonPackages := []string{
		"git",
		"curl",
		"wget",
		"tmux",
		"vim",
		"htop",
		"tree",
		"jq",
		"ripgrep",
		"fzf",
	}

	platformSpecificPackages := map[platform.Platform][]string{
		platform.MacOS: {
			"coreutils",
			"gnused",
			"gawk",
			"findutils",
			"gnu-tar",
			"gnutls",
			"gnupg",
			"bash",
			"zsh",
			"fish",
		},
		platform.Linux: {
			"gcc",
			"make",
			"openssh",
			"rsync",
			"unzip",
			"zip",
			"gzip",
			"tar",
			"which",
			"xz",
		},
	}

	var packages []string
	if m.isWSL {
		packages = append(packages, platformSpecificPackages[platform.Linux]...)
	} else {
		packages = append(packages, platformSpecificPackages[m.platform]...)
	}

	packages = append(packages, commonPackages...)

	return packages
}

// ValidatePackage checks if a package is valid for the current platform.
func (m *Manager) ValidatePackage(pkg string) error {
	invalidPackages := map[platform.Platform][]string{
		platform.MacOS: {
			"systemd",
			"apt",
			"dnf",
			"yum",
			"pacman",
		},
		platform.Linux: {
			"xcode",
			"launchctl",
			"brew",
			"port",
		},
	}

	if m.isWSL {
		for _, invalidPkg := range invalidPackages[platform.Linux] {
			if pkg == invalidPkg {
				return fmt.Errorf("package %s is not available on WSL", pkg)
			}
		}
	} else {
		for _, invalidPkg := range invalidPackages[m.platform] {
			if pkg == invalidPkg {
				return fmt.Errorf("package %s is not available on %s", pkg, m.platform)
			}
		}
	}

	return nil
}

// GetPackageGroups returns predefined groups of packages for common use cases.
func (m *Manager) GetPackageGroups() map[string][]string {
	return map[string][]string{
		"development": {
			"git",
			"gcc",
			"make",
			"python3",
			"nodejs",
			"yarn",
			"docker",
			"docker-compose",
			"kubernetes-cli",
			"terraform",
		},
		"system": {
			"htop",
			"tree",
			"tmux",
			"vim",
			"neovim",
			"rsync",
			"curl",
			"wget",
			"jq",
			"ripgrep",
		},
		"security": {
			"gnupg",
			"openssh",
			"openssl",
			"vault",
			"age",
			"sops",
			"yubikey-manager",
			"pass",
			"gopass",
			"keychain",
		},
		"shell": {
			"zsh",
			"fish",
			"starship",
			"fzf",
			"bat",
			"exa",
			"fd",
			"direnv",
			"zoxide",
			"thefuck",
		},
	}
}

// GetPackageDescription returns a description of what a package does.
func (m *Manager) GetPackageDescription(pkg string) string {
	descriptions := map[string]string{
		"git":             "Distributed version control system",
		"gcc":             "GNU Compiler Collection",
		"make":            "Build automation tool",
		"python3":         "Python programming language interpreter",
		"nodejs":          "JavaScript runtime",
		"yarn":            "Fast, reliable, and secure dependency management",
		"docker":          "Container platform",
		"docker-compose":  "Tool for defining and running multi-container Docker applications",
		"kubernetes-cli":  "Kubernetes command-line tool",
		"terraform":       "Infrastructure as code software tool",
		"htop":            "Interactive process viewer",
		"tree":            "Directory listing tool",
		"tmux":            "Terminal multiplexer",
		"vim":             "Highly configurable text editor",
		"neovim":          "Hyperextensible Vim-based text editor",
		"rsync":           "Fast, versatile file copying tool",
		"curl":            "Command line tool for transferring data",
		"wget":            "Network utility to retrieve files from the web",
		"jq":              "Lightweight command-line JSON processor",
		"ripgrep":         "Fast line-oriented search tool",
		"gnupg":           "GNU Privacy Guard",
		"openssh":         "OpenBSD Secure Shell",
		"openssl":         "Cryptography and SSL/TLS toolkit",
		"vault":           "Tool for secrets management and data protection",
		"age":             "Simple, modern file encryption tool",
		"sops":            "Secrets management tool",
		"yubikey-manager": "Tool for managing YubiKey security keys",
		"pass":            "Standard Unix password manager",
		"gopass":          "Team password manager using git",
		"keychain":        "SSH and GPG agent management",
		"zsh":             "Z shell with lots of features",
		"fish":            "User-friendly command line shell",
		"starship":        "Cross-shell prompt",
		"fzf":             "Command-line fuzzy finder",
		"bat":             "Cat clone with syntax highlighting",
		"exa":             "Modern replacement for ls",
		"fd":              "Simple, fast alternative to find",
		"direnv":          "Environment switcher for the shell",
		"zoxide":          "Smarter cd command",
		"thefuck":         "Magnificent app which corrects your previous console command",
	}

	if desc, ok := descriptions[pkg]; ok {
		return desc
	}
	return "No description available"
}
