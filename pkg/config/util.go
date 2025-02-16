package config

import (
	"os/exec"
	"strings"
)

// GetGitConfig retrieves Git configuration values
func GetGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
