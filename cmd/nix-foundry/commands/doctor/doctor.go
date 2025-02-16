package doctor

import (
	"fmt"
	"os"

	"github.com/shawnkhoffman/nix-foundry/internal/services/health"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewDoctorCommand(projectSvc project.Service) *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose environment and configuration issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			spin := progress.NewSpinner("Running system diagnostics...")
			spin.Start()

			// Run all health checks
			systemChecks := health.RunSystemChecks()
			configChecks := health.NewConfigChecker(projectSvc).AuditConfigs()

			spin.Stop()

			// Display results
			fmt.Println("\nSystem Health Report:")
			printResults(systemChecks, verbose)

			fmt.Println("\nConfiguration Audit:")
			printResults(configChecks, verbose)

			// Exit code handling
			if hasCriticalErrors(systemChecks, configChecks) {
				fmt.Fprintln(os.Stderr, "\n❌ Critical issues detected")
				os.Exit(1)
			}

			fmt.Println("\n✅ System appears healthy")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed diagnostic info")
	return cmd
}

func printResults(checks []health.CheckResult, verbose bool) {
	for _, result := range checks {
		status := "✅"
		if result.Status == health.StatusWarning {
			status = "⚠️"
		} else if result.Status == health.StatusError {
			status = "❌"
		}

		fmt.Printf("  %s %s", status, result.Name)
		if verbose {
			fmt.Printf("\n    %s\n", result.Details)
		} else {
			fmt.Println()
		}
	}
}

func hasCriticalErrors(checkGroups ...[]health.CheckResult) bool {
	for _, group := range checkGroups {
		for _, check := range group {
			if check.Status == health.StatusError {
				return true
			}
		}
	}
	return false
}
