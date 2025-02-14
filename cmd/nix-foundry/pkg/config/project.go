package config

import (
	"fmt"
)

// LoadProjectConfig loads and merges project configuration with team config if specified
func (cm *Manager) LoadProjectConfig(teamName string) (*ProjectConfig, error) {
	var project ProjectConfig

	// Load base project config
	if err := cm.ReadConfig("project.yaml", &project); err != nil {
		return nil, err
	}

	// If team config specified, merge it
	if teamName != "" {
		team, err := cm.loadTeamConfig(teamName)
		if err != nil {
			return nil, fmt.Errorf("failed to load team config: %w", err)
		}
		merged := cm.MergeProjectConfigs(project, *team)
		return &merged, nil
	}

	return &project, nil
}

// LoadProjectWithTeam loads project configuration and merges with team config if specified
func (cm *Manager) LoadProjectWithTeam(projectName, teamName string) (*ProjectConfig, error) {
	project, err := cm.LoadConfig(ProjectConfigType, projectName)
	if err != nil {
		return nil, err
	}

	if teamName == "" {
		return project.(*ProjectConfig), nil
	}

	team, err := cm.LoadConfig(TeamConfigType, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}

	projectConfig, ok := project.(*ProjectConfig)
	if !ok {
		return nil, fmt.Errorf("invalid project configuration type")
	}

	teamConfig, ok := team.(*ProjectConfig)
	if !ok {
		return nil, fmt.Errorf("invalid team configuration type")
	}

	merged := cm.MergeProjectConfigs(*projectConfig, *teamConfig)
	return &merged, nil
}
