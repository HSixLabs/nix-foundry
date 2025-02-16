package status

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/status"
	"github.com/spf13/cobra"
)

func NewEnvironmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environment",
		Short: "Show environment status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize required services
			configSvc := config.NewService()
			validator := validation.NewService()
			platformSvc := platform.NewService()

			envSvc := environment.NewService(
				configSvc.GetConfigDir(),
				configSvc,
				validator,
				platformSvc,
			)

			svc := status.NewService(configSvc, envSvc)
			status, err := svc.CheckEnvironment()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ACTIVE ENVIRONMENT:\t%s\n", status.Active)
			fmt.Fprintf(w, "LAST APPLIED:\t%s\n", status.LastApply.Format(time.RFC1123))
			fmt.Fprintf(w, "PACKAGE COUNT:\t%d\n", len(status.Packages))
			fmt.Fprintf(w, "HEALTH STATUS:\t%s\n", status.Health)
			w.Flush()

			return nil
		},
	}
	return cmd
}
