package flags

import "github.com/spf13/cobra"

const (
	ForceFlag   = "force"
	YesFlag     = "yes"
	VerboseFlag = "verbose"
	ConfigFlag  = "config"
)

// AddGlobalFlags adds flags that persist through all commands
func AddGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(VerboseFlag, "v", false,
		"Enable verbose output")
	cmd.PersistentFlags().String(ConfigFlag, "",
		"Config file (default $HOME/.nix-foundry/config.yaml)")
}

// AddForceFlag adds --force to a command
func AddForceFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP(ForceFlag, "f", false,
		"Force operation despite conflicts")
}

// AddYesFlag adds --yes for automatic confirmation
func AddYesFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP(YesFlag, "y", false,
		"Automatic yes to prompts")
}
