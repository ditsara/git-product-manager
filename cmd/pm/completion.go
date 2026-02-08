package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for pm.

To load completions:

Bash:
  $ source <(pm completion bash)
  
  # To load completions for each session, add to ~/.bashrc:
  $ pm completion bash > ~/.pm-completion.sh
  $ echo 'source ~/.pm-completion.sh' >> ~/.bashrc

Zsh:
  $ source <(pm completion zsh)
  
  # To load completions for each session, add to ~/.zshrc:
  $ pm completion zsh > "${fpath[1]}/_pm"

Fish:
  $ pm completion fish | source
  
  # To load completions for each session:
  $ pm completion fish > ~/.config/fish/completions/pm.fish

PowerShell:
  PS> pm completion powershell | Out-String | Invoke-Expression
  
  # To load completions for each session, add to your PowerShell profile:
  PS> pm completion powershell >> $PROFILE
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
