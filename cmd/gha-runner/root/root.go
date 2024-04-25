package root

import (
	"github.com/dungdm93/drasi/cmd/gha-runner/configure"
	"github.com/spf13/cobra"
)

func NewGHARunnerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gha-runner",
		Short: "GitHub Actions runner (re-)implemented in Go",
		Args:  cobra.NoArgs,
	}

	AddCommands(cmd)

	return cmd
}

func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		configure.NewConfigureCommand(),
	)
}
