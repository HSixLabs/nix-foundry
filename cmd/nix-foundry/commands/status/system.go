package status

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/status"
	"github.com/spf13/cobra"
)

func NewSystemCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "system",
		Short: "Show system status",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := status.NewService(config.NewService(), nil)
			status, err := svc.CheckSystem()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "NIX VERSION:\t%s\n", status.NixVersion)
			fmt.Fprintf(w, "STORAGE:\n%s\n", status.Storage)
			fmt.Fprintf(w, "DEPENDENCIES:\t%v\n", status.Dependencies)
			fmt.Fprintf(w, "SERVICES:\t%v\n", status.ServiceStatus)
			w.Flush()

			return nil
		},
	}
}
