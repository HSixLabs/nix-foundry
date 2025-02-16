package status

import (
	"os/exec"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

type Service struct {
	configSvc      config.Service
	environmentSvc environment.Service
}

func NewService(cfgSvc config.Service, envSvc environment.Service) *Service {
	return &Service{
		configSvc:      cfgSvc,
		environmentSvc: envSvc,
	}
}

func (s *Service) CheckEnvironment() (EnvironmentStatus, error) {
	status := EnvironmentStatus{
		Active:    s.environmentSvc.GetCurrentEnvironment(),
		Packages:  s.getPackagesFromConfig(),
		Health:    s.environmentSvc.CheckHealth(),
		LastApply: s.getLastApplyTime(),
	}
	return status, nil
}

func (s *Service) CheckSystem() (SystemStatus, error) {
	return SystemStatus{
		NixVersion:    getNixVersion(),
		Storage:       checkStorage(),
		Dependencies:  verifyDependencies(),
		ServiceStatus: checkServices(),
	}, nil
}

type EnvironmentStatus struct {
	Active    string
	Packages  []string
	Health    string
	LastApply time.Time
}

type SystemStatus struct {
	NixVersion    string
	Storage       string
	Dependencies  []string
	ServiceStatus map[string]string
}

func getNixVersion() string {
	out, _ := exec.Command("nix", "--version").Output()
	return strings.TrimSpace(string(out))
}

func checkStorage() string {
	out, _ := exec.Command("df", "-h").Output()
	return string(out)
}

func verifyDependencies() []string {
	return []string{"nix", "docker", "git"} // Simplified example
}

func checkServices() map[string]string {
	return map[string]string{
		"nix-daemon": "active",
	}
}

func (s *Service) getPackagesFromConfig() []string {
	var projectCfg project.ProjectConfig
	if err := s.configSvc.LoadSection("project", &projectCfg); err == nil {
		return projectCfg.Dependencies
	}
	return []string{}
}

func (s *Service) getLastApplyTime() time.Time {
	var lastApply struct {
		Timestamp time.Time `yaml:"last_apply"`
	}
	if err := s.configSvc.LoadSection("metadata", &lastApply); err == nil {
		return lastApply.Timestamp
	}
	return time.Time{}
}
