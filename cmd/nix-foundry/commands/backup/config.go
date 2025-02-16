package backup

import (
	"fmt"

	backupsvc "github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

func newConfigCmd(svc backupsvc.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage backup configuration",
	}

	cmd.AddCommand(
		newConfigShowCmd(svc),
		newConfigSetCmd(svc),
	)

	return cmd
}

func newConfigShowCmd(svc backupsvc.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := svc.GetConfig()
			fmt.Printf("Backup Configuration:\n")
			fmt.Printf("Retention: %d days\n", config.RetentionDays)
			fmt.Printf("Max Backups: %d\n", config.MaxBackups)
			fmt.Printf("Compression Level: %d\n", config.CompressionLevel)
			return nil
		},
	}
}

func newConfigSetCmd(svc backupsvc.Service) *cobra.Command {
	var (
		retentionDays    int
		maxBackups       int
		compressionLevel int
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := svc.GetConfig()

			if cmd.Flag("retention").Changed {
				config.RetentionDays = retentionDays
			}
			if cmd.Flag("max-backups").Changed {
				config.MaxBackups = maxBackups
			}
			if cmd.Flag("compression").Changed {
				config.CompressionLevel = compressionLevel
			}

			return svc.UpdateConfig(config)
		},
	}

	cmd.Flags().IntVar(&retentionDays, "retention", 0, "Days to keep backups")
	cmd.Flags().IntVar(&maxBackups, "max-backups", 0, "Maximum number of backups")
	cmd.Flags().IntVar(&compressionLevel, "compression", 0, "Compression level (1-9)")

	return cmd
}

func NewSetConfigCmd(svc backupsvc.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "set-config [key] [value]",
		Short: "Set backup configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Convert key/value pair to Config structure
			config := svc.GetConfig()

			switch args[0] {
			case "retention":
				var days int
				if _, err := fmt.Sscan(args[1], &days); err != nil {
					return fmt.Errorf("invalid retention days: %w", err)
				}
				config.RetentionDays = days
			case "max-backups":
				var max int
				if _, err := fmt.Sscan(args[1], &max); err != nil {
					return fmt.Errorf("invalid max backups: %w", err)
				}
				config.MaxBackups = max
			case "compression":
				var level int
				if _, err := fmt.Sscan(args[1], &level); err != nil {
					return fmt.Errorf("invalid compression level: %w", err)
				}
				config.CompressionLevel = level
			default:
				return fmt.Errorf("unknown config key: %s", args[0])
			}

			return svc.UpdateConfig(config)
		},
	}
}
