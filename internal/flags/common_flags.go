package flags

import "github.com/spf13/cobra"

// Single source for test mode flag
func AddTestModeFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().BoolP("test", "t", false, "Run in test mode")
}
