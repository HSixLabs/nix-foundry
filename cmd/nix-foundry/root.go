package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nix-foundry",
	Short: "A dual-purpose environment manager",
	Long: `nix-foundry is a dual-purpose environment manager that handles both
personal development setups and team project environments using Nix.`,
	SilenceErrors: true,
}

func init() {
	// Register each command exactly once
	rootCmd.AddCommand(
		applyCmd,
		backupCmd,
		configCmd,
		doctorCmd,
		initCmd,
		packagesCmd,
		profileCmd,
		projectCmd,
		rollbackCmd,
		statusCmd,
		switchCmd,
		uninstallCmd,
		updateCmd,
		newCompletionCmd(),
	)

	// Set up command flags
	doctorCmd.Flags().BoolVar(&security, "security", false, "Run security audit")
	doctorCmd.Flags().BoolVar(&system, "system", false, "Run system checks only")
	doctorCmd.Flags().BoolVar(&fix, "fix", false, "Attempt to fix issues")
	doctorCmd.Flags().BoolVar(&follow, "follow", false, "Follow log output")
	doctorCmd.Flags().IntVar(&tail, "tail", 10, "Number of log lines to show")
}

func Execute() error {
	return rootCmd.Execute()
}

// Add this function to generate the completion command
func newCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `To load completions:

Bash:
  $ source <(nix-foundry completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ nix-foundry completion bash > /etc/bash_completion.d/nix-foundry
  # macOS:
  $ nix-foundry completion bash > $(brew --prefix)/etc/bash_completion.d/nix-foundry

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ nix-foundry completion zsh > "${fpath[1]}/_nix-foundry"

Fish:
  $ nix-foundry completion fish | source

  # To load completions for each session, execute once:
  $ nix-foundry completion fish > ~/.config/fish/completions/nix-foundry.fish

PowerShell:
  PS> nix-foundry completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> nix-foundry completion powershell > nix-foundry.ps1
  # and source this file from your PowerShell profile.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell type %q", args[0])
			}
		},
	}
	return completionCmd
}
