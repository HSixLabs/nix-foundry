package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

func NewCheckConflictsCommand(projectSvc project.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-conflicts",
		Short: "Check for configuration conflicts",
		Long:  `Check for conflicts between personal and team configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := projectSvc.ValidateConflicts(nil); err != nil {
				return fmt.Errorf("configuration conflicts detected: %w", err)
			}

			fmt.Println("Configuration validation successful - no conflicts")
			return nil
		},
	}

	return cmd
}
