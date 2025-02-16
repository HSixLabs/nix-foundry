package health

import (
	"os"
	"os/exec"
	"path/filepath"
)

type SystemChecker struct{}

func (s *SystemChecker) RunChecks() []CheckResult {
	return []CheckResult{
		s.checkNixInstalled(),
		s.checkDiskSpace(),
		s.checkNetworkAccess(),
		s.checkEnvPermissions(),
	}
}

func (s *SystemChecker) checkNixInstalled() CheckResult {
	_, err := exec.LookPath("nix")
	errorMsg := ""
	if err != nil {
		errorMsg = "Nix package manager not found in PATH"
	}
	return CheckResult{
		Name:    "Nix Installation",
		Status:  mapBoolToStatus(err == nil),
		Details: errorMsg,
	}
}

func (s *SystemChecker) checkDiskSpace() CheckResult {
	return CheckResult{
		Name:   "Disk Space",
		Status: StatusOK, // TODO: Implement actual check
	}
}

func (s *SystemChecker) checkNetworkAccess() CheckResult {
	return CheckResult{
		Name:   "Network Access",
		Status: StatusOK, // TODO: Implement actual check
	}
}

func (s *SystemChecker) checkEnvPermissions() CheckResult {
	home := os.Getenv("HOME")
	status := true
	var errorMsg string

	if fi, err := os.Stat(filepath.Join(home, ".nix")); err == nil {
		if fi.Mode().Perm()&0200 == 0 {
			status = false
			errorMsg = "Missing write permissions for .nix directory"
		}
	}

	return CheckResult{
		Name:    "Environment Permissions",
		Status:  mapBoolToStatus(status),
		Details: errorMsg,
	}
}

func RunSystemChecks() []CheckResult {
	return []CheckResult{
		checkNixInstallation(),
		checkFlakeSupport(),
		checkDiskSpace(),
	}
}

func checkNixInstallation() CheckResult {
	_, err := exec.LookPath("nix")
	if err != nil {
		return CheckResult{
			Name:    "Nix Installation",
			Status:  StatusError,
			Details: "Nix package manager not found in PATH",
		}
	}
	return CheckResult{
		Name:   "Nix Installation",
		Status: StatusOK,
	}
}

func checkFlakeSupport() CheckResult {
	// Temporary implementation
	return CheckResult{
		Name:   "Flake Support",
		Status: StatusOK,
	}
}

func checkDiskSpace() CheckResult {
	// Temporary implementation
	return CheckResult{
		Name:   "Disk Space",
		Status: StatusOK,
	}
}

func mapBoolToStatus(ok bool) CheckStatus {
	if ok {
		return StatusOK
	}
	return StatusError
}
