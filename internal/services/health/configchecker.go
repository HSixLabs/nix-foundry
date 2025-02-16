package health

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

type CheckResult struct {
	Name    string
	Status  CheckStatus
	Details string
}

type ConfigChecker struct {
	projectSvc project.Service
}

func NewConfigChecker(svc project.Service) *ConfigChecker {
	return &ConfigChecker{projectSvc: svc}
}

func (c *ConfigChecker) AuditConfigs() []CheckResult {
	var results []CheckResult

	results = append(results, c.checkProjectConfig())
	results = append(results, c.checkEnvironmentConfig())
	results = append(results, c.checkTeamConfigs())

	return results
}

func (c *ConfigChecker) checkProjectConfig() CheckResult {
	config := c.projectSvc.GetProjectConfig()

	if config == nil {
		return CheckResult{
			Name:    "Project Config Validity",
			Status:  StatusError,
			Details: "Project config not loaded",
		}
	}

	if err := config.Validate(); err != nil {
		return CheckResult{
			Name:    "Project Config Validity",
			Status:  StatusError,
			Details: fmt.Sprintf("Invalid project config: %v", err),
		}
	}

	return CheckResult{
		Name:   "Project Config Validity",
		Status: StatusOK,
	}
}

func (c *ConfigChecker) checkEnvironmentConfig() CheckResult {
	config := c.projectSvc.GetProjectConfig()
	if config == nil {
		return CheckResult{
			Name:    "Environment Config",
			Status:  StatusError,
			Details: "Project config not loaded",
		}
	}

	// Check environment isolation directory exists
	envPath := filepath.Join(c.projectSvc.GetConfigDir(), "environments", config.Environment)
	_, err := os.Stat(envPath)

	if os.IsNotExist(err) {
		return CheckResult{
			Name:    "Environment Configuration",
			Status:  StatusError,
			Details: fmt.Sprintf("Environment directory %s not found", envPath),
		}
	} else if err != nil {
		return CheckResult{
			Name:    "Environment Configuration",
			Status:  StatusError,
			Details: fmt.Sprintf("Error checking environment: %v", err),
		}
	}

	return CheckResult{
		Name:   "Environment Configuration",
		Status: StatusOK,
	}
}

func (c *ConfigChecker) checkTeamConfigs() CheckResult {
	config := c.projectSvc.GetProjectConfig()
	if config == nil {
		return CheckResult{
			Name:    "Team Configs",
			Status:  StatusError,
			Details: "Project config not loaded",
		}
	}

	// Check for required team settings
	if len(config.Dependencies) == 0 {
		return CheckResult{
			Name:    "Team Configuration",
			Status:  StatusError,
			Details: "No team dependencies configured",
		}
	}

	// Validate dependency format
	for _, dep := range config.Dependencies {
		if strings.TrimSpace(dep) == "" {
			return CheckResult{
				Name:    "Team Configuration",
				Status:  StatusError,
				Details: "Empty dependency in team configuration",
			}
		}
	}

	return CheckResult{
		Name:   "Team Configuration",
		Status: StatusOK,
	}
}
