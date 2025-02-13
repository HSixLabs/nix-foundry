package platform

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// System represents the current platform's system information
type System struct {
	OS    string
	Arch  string
	IsWSL bool
}

// Detect returns information about the current system platform
func Detect() (*System, error) {
	sys := &System{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	// Validate supported platforms
	switch sys.OS {
	case "darwin", "linux":
		// These platforms are supported
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", sys.OS)
	}

	// Validate supported architectures
	switch sys.Arch {
	case "amd64", "arm64":
		// These architectures are supported
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", sys.Arch)
	}

	// Check for WSL
	if sys.OS == "linux" {
		if data, err := os.ReadFile("/proc/version"); err == nil {
			sys.IsWSL = strings.Contains(strings.ToLower(string(data)), "microsoft")
		}
	}

	return sys, nil
}
