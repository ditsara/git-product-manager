package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for pm.

Add this to your shell config for persistent completions:

Bash (~/.bashrc):
  source <(pm completion bash)

Zsh (~/.zshrc):
  source <(pm completion zsh)

Fish (~/.config/fish/config.fish):
  pm completion fish | source

PowerShell ($PROFILE):
  pm completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			cmd.PrintErrf("Error: unknown shell %q. Valid options: bash, zsh, fish, powershell\n", args[0])
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
