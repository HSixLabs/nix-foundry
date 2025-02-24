package set

import (
	"github.com/spf13/cobra"
)

// Cmd represents the set command
var Cmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long: `Set configuration values.
This command provides subcommands for setting various configuration values.`,
}
