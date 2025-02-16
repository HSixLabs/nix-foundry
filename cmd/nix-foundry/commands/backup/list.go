package backup

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

func NewListCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available backups",
		Long: `List all available nix-foundry configuration backups.
Shows backup name, creation time, and size.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backups, err := svc.List()
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			if len(backups) == 0 {
				fmt.Println("No backups found")
				return nil
			}

			// Initialize tabwriter
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			defer w.Flush()

			fmt.Fprintln(w, "ID\tTIMESTAMP\tSIZE\tPATH")
			for _, b := range backups {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
					b.ID,
					b.Timestamp.Format("2006-01-02 15:04"),
					b.Size,
					b.Path)
			}

			return nil
		},
	}
}
