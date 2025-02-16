package flags

import "github.com/spf13/cobra"

func AddTestModeFlag(cmd *cobra.Command) *bool {
	return cmd.Flags().BoolP("test", "t", false, "Run in test mode")
}

func AddVerboseFlag(cmd *cobra.Command, verbose *bool) {
	cmd.Flags().BoolVarP(verbose, "verbose", "v", false, "Show detailed output")
}
