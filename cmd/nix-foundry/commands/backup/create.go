package backup

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewCreateCmd(svc backup.Service) *cobra.Command {
	var (
		force    bool
		compress bool
	)

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create new backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := time.Now().Format("20060102-150405")
			if len(args) > 0 {
				name = args[0]
			}

			spin := progress.NewSpinner("Creating backup...")
			spin.Start()

			if err := svc.Create(name, force); err != nil {
				spin.Fail("Backup creation failed")
				return fmt.Errorf("backup creation failed: %w", err)
			}

			if compress {
				spin.Update("Compressing backup...")
				if err := svc.Compress(name); err != nil {
					spin.Fail("Compression failed")
					return fmt.Errorf("backup compression failed: %w", err)
				}
			}

			spin.Success("Backup created")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing backup")
	cmd.Flags().BoolVarP(&compress, "compress", "c", true, "Enable compression")

	return cmd
}
