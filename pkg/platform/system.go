package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Unified System struct with WSL detection
type System struct {
	OS    string
	Arch  string
	IsWSL bool
}

// Detect returns the current system configuration
func Detect() (*System, error) {
	sys := &System{
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
		IsWSL: detectWSL(),
	}

	// Validation remains the same
	if err := validatePlatform(sys.OS, sys.Arch); err != nil {
		return nil, err
	}

	return sys, nil
}

// Add explicit validation function
func validatePlatform(os, arch string) error {
	switch os {
	case "darwin", "linux":
	default:
		return fmt.Errorf("unsupported OS: %s", os)
	}

	switch arch {
	case "amd64", "arm64":
	default:
		return fmt.Errorf("unsupported architecture: %s", arch)
	}
	return nil
}

// Enhanced WSL detection
func detectWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// Check multiple WSL indicators
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSL"); err == nil {
		return true
	}

	if data, err := os.ReadFile("/proc/version"); err == nil {
		return strings.Contains(strings.ToLower(string(data)), "microsoft")
	}

	return false
}

// NewSystem detects current platform characteristics
func NewSystem() *System {
	return &System{
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
		IsWSL: detectWSL(),
	}
}
