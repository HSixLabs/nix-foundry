package health

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type SystemChecker struct{}

func (s *SystemChecker) RunChecks() []CheckResult {
	return []CheckResult{
		s.checkNixInstalled(),
		s.checkDiskSpace(),
		s.checkNetworkAccess(),
		s.checkEnvPermissions(),
		s.checkFlakeSupport(),
	}
}

func (s *SystemChecker) checkNixInstalled() CheckResult {
	path, err := exec.LookPath("nix")
	var details string
	if err != nil {
		details = "Nix package manager not found in PATH"
	} else {
		// Additional version check
		cmd := exec.Command("nix", "--version")
		if output, err := cmd.Output(); err == nil {
			details = "Nix installed at: " + path + "\nVersion: " + string(output)
		}
	}
	return CheckResult{
		Name:    "Nix Installation",
		Status:  mapBoolToStatus(err == nil),
		Details: details,
	}
}

func (s *SystemChecker) checkDiskSpace() CheckResult {
	home := os.Getenv("HOME")
	var stat syscall.Statfs_t
	err := syscall.Statfs(home, &stat)

	if err != nil {
		return CheckResult{
			Name:    "Disk Space",
			Status:  StatusWarning,
			Details: "Unable to check disk space",
		}
	}

	// Calculate available space in GB
	available := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024 * 1024)

	if available < 10 { // Less than 10GB
		return CheckResult{
			Name:    "Disk Space",
			Status:  StatusWarning,
			Details: fmt.Sprintf("Low disk space: %dGB available", available),
		}
	}

	return CheckResult{
		Name:    "Disk Space",
		Status:  StatusOK,
		Details: fmt.Sprintf("Available space: %dGB", available),
	}
}

func (s *SystemChecker) checkNetworkAccess() CheckResult {
	// Try to connect to common Nix endpoints
	endpoints := []string{
		"cache.nixos.org:443",
		"github.com:443",
	}

	var failedEndpoints []string
	for _, endpoint := range endpoints {
		conn, err := net.DialTimeout("tcp", endpoint, 5*time.Second)
		if err != nil {
			failedEndpoints = append(failedEndpoints, endpoint)
		} else {
			conn.Close()
		}
	}

	if len(failedEndpoints) > 0 {
		return CheckResult{
			Name:    "Network Access",
			Status:  StatusWarning,
			Details: fmt.Sprintf("Cannot connect to: %v", failedEndpoints),
		}
	}

	return CheckResult{
		Name:    "Network Access",
		Status:  StatusOK,
		Details: "All required endpoints are accessible",
	}
}

func (s *SystemChecker) checkEnvPermissions() CheckResult {
	home := os.Getenv("HOME")
	var details []string
	status := true

	// Check .nix directory
	nixPath := filepath.Join(home, ".nix")
	if fi, err := os.Stat(nixPath); err == nil {
		if fi.Mode().Perm()&0200 == 0 {
			status = false
			details = append(details, "Missing write permissions for .nix directory")
		}
	}

	// Check nix store permissions
	nixStore := "/nix/store"
	if fi, err := os.Stat(nixStore); err == nil {
		if fi.Mode().Perm()&0400 == 0 {
			status = false
			details = append(details, "Missing read permissions for /nix/store")
		}
	}

	return CheckResult{
		Name:    "Environment Permissions",
		Status:  mapBoolToStatus(status),
		Details: strings.Join(details, "\n"),
	}
}

func (s *SystemChecker) checkFlakeSupport() CheckResult {
	cmd := exec.Command("nix", "flake", "--help")
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:    "Flake Support",
			Status:  StatusError,
			Details: "Nix flakes not enabled. Enable with 'nix-env -iA nixpkgs.nixFlakes'",
		}
	}

	// Check experimental features
	cmd = exec.Command("nix", "show-config")
	output, err := cmd.Output()
	if err == nil && !strings.Contains(string(output), "experimental-features") {
		return CheckResult{
			Name:    "Flake Support",
			Status:  StatusWarning,
			Details: "Flakes available but experimental features may not be configured",
		}
	}

	return CheckResult{
		Name:    "Flake Support",
		Status:  StatusOK,
		Details: "Nix flakes enabled and configured",
	}
}

func mapBoolToStatus(ok bool) CheckStatus {
	if ok {
		return StatusOK
	}
	return StatusError
}
