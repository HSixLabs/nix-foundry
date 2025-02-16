package environment

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/flags"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/spf13/cobra"
)

func NewRollbackCommand(svc environment.Service) *cobra.Command {
	var (
		timestamp string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "rollback [timestamp|duration]",
		Short: "Rollback to a previous environment state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target, err := parseTimestamp(args[0])
			if err != nil {
				return fmt.Errorf("invalid timestamp: %w", err)
			}

			if err := svc.Rollback(target, force); err != nil {
				if !force {
					fmt.Printf("Rollback failed: %v\nUse --force to override conflicts", err)
					return err
				}
				return fmt.Errorf("forced rollback failed: %w", err)
			}

			fmt.Printf("Successfully rolled back to %s\n", target.Format(time.RFC3339))
			return nil
		},
	}

	flags.AddForceFlag(cmd)
	cmd.Flags().StringVarP(&timestamp, "timestamp", "t", "", "Exact backup timestamp (YYYYMMDD-HHMMSS)")
	return cmd
}

func parseTimestamp(input string) (time.Time, error) {
	// Try duration format first
	if duration, err := time.ParseDuration(input); err == nil {
		return time.Now().Add(-duration), nil
	}

	// Try timestamp format
	return time.Parse("20060102-150405", input)
}
