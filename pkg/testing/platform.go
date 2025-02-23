// Package testing provides testing utilities.
package testing

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/platform"
)

// PlatformTest provides utilities for platform-specific testing
type PlatformTest struct {
	platform platform.Platform
	isWSL    bool
	tempDir  string
	results  []TestResult
}

// NewPlatformTest creates a new platform test utility
func NewPlatformTest() (*PlatformTest, error) {
	tempDir, err := os.MkdirTemp("", "nix-foundry-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &PlatformTest{
		platform: platform.GetPlatform(),
		isWSL:    platform.IsWSL(),
		tempDir:  tempDir,
		results:  make([]TestResult, 0),
	}, nil
}

// Cleanup removes temporary test files
func (p *PlatformTest) Cleanup() error {
	return os.RemoveAll(p.tempDir)
}

// CreateTestConfig creates a test configuration file
func (p *PlatformTest) CreateTestConfig() (string, error) {
	configPath := filepath.Join(p.tempDir, "test-config.yaml")
	content := `version: "1.0"
kind: test
metadata:
  name: test-config
settings:
  autoUpdate: true
  updateInterval: 24h
  logLevel: info
nix:
  packages:
    core:
      - git
      - curl
    optional:
      - vim
      - tmux
  shell: /bin/zsh
  manager: nix-env
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write test config: %w", err)
	}
	return configPath, nil
}

// CreateTestScript creates a test shell script
func (p *PlatformTest) CreateTestScript() (string, error) {
	scriptPath := filepath.Join(p.tempDir, "test-script.sh")
	content := `#!/usr/bin/env bash
echo "Test script running on $(uname -s)"
exit 0
`
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		return "", fmt.Errorf("failed to write test script: %w", err)
	}
	return scriptPath, nil
}

// CreateTestEnvironment sets up a test environment with shell configuration
func (p *PlatformTest) CreateTestEnvironment() error {
	shellConfigDir := filepath.Join(p.tempDir, ".config", "shell")
	if err := os.MkdirAll(shellConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create shell config directory: %w", err)
	}

	shellConfig := filepath.Join(shellConfigDir, "config")
	content := `# Test shell configuration
PATH="/test/bin:$PATH"
SHELL="/bin/test-shell"
`
	if err := os.WriteFile(shellConfig, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write shell config: %w", err)
	}

	return nil
}

// RunPlatformTests runs platform-specific tests
func (p *PlatformTest) RunPlatformTests() error {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"TestFileSystem", p.testFileSystem},
		{"TestShellConfig", p.testShellConfig},
		{"TestPackageManager", p.testPackageManager},
		{"TestPlatformPaths", p.testPlatformPaths},
	}

	for _, test := range tests {
		result := TestResult{
			Name:      test.name,
			StartTime: time.Now(),
		}

		fmt.Printf("Running %s...\n", test.name)
		if err := test.fn(); err != nil {
			result.Error = err
			result.Passed = false
			result.EndTime = time.Now()
			p.results = append(p.results, result)
			return fmt.Errorf("%s failed: %w", test.name, err)
		}

		result.Passed = true
		result.EndTime = time.Now()
		p.results = append(p.results, result)
		fmt.Printf("✓ %s passed\n", test.name)
	}

	return nil
}

func (p *PlatformTest) testFileSystem() error {
	testFile := filepath.Join(p.tempDir, "test.txt")
	content := "test content"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}
	if string(data) != content {
		return fmt.Errorf("file content mismatch: got %q, want %q", string(data), content)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		return fmt.Errorf("failed to stat test file: %w", err)
	}
	if info.Mode().Perm() != 0644 {
		return fmt.Errorf("file permissions mismatch: got %v, want %v", info.Mode().Perm(), 0644)
	}

	return nil
}

func (p *PlatformTest) testShellConfig() error {
	var shellPath string
	switch {
	case p.isWSL:
		shellPath = "/bin/bash"
	case p.platform == platform.MacOS:
		shellPath = "/bin/zsh"
	default:
		shellPath = "/bin/bash"
	}

	if _, err := os.Stat(shellPath); err != nil {
		return fmt.Errorf("shell not found at %s: %w", shellPath, err)
	}

	cmd := exec.Command(shellPath, "-c", "echo $SHELL")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute shell command: %w", err)
	}
	if !strings.Contains(string(output), shellPath) {
		return fmt.Errorf("shell path mismatch: got %q, want %q", string(output), shellPath)
	}

	return nil
}

func (p *PlatformTest) testPackageManager() error {
	if _, err := exec.LookPath("nix-env"); err != nil {
		return fmt.Errorf("nix-env not found in PATH: %w", err)
	}

	cmd := exec.Command("nix-env", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get nix-env version: %w", err)
	}

	return nil
}

func (p *PlatformTest) testPlatformPaths() error {
	paths := map[string]string{
		"HOME": os.Getenv("HOME"),
		"PATH": os.Getenv("PATH"),
		"PWD":  os.Getenv("PWD"),
	}

	for name, value := range paths {
		if value == "" {
			return fmt.Errorf("environment variable %s is not set", name)
		}
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(os.Getenv("PATH"), ";") {
			return fmt.Errorf("PATH separator is not ; on Windows")
		}
	} else {
		if !strings.Contains(os.Getenv("PATH"), ":") {
			return fmt.Errorf("PATH separator is not : on Unix")
		}
	}

	return nil
}

// GetTempDir returns the temporary directory path
func (p *PlatformTest) GetTempDir() string {
	return p.tempDir
}

// GetPlatform returns the current platform
func (p *PlatformTest) GetPlatform() platform.Platform {
	return p.platform
}

// GetTestResults returns the results of all tests
func (p *PlatformTest) GetTestResults() []TestResult {
	return p.results
}

// IsWSL returns whether running under WSL
func (p *PlatformTest) IsWSL() bool {
	return p.isWSL
}
