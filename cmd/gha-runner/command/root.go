package command

import (
	"github.com/spf13/cobra"
)

func NewGHARunnerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gha-runner",
		Short: "GitHub Actions runner (re-)implemented in Go",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		NewRegisterCommand(),
		NewLaunchCommand(),
	)

	return cmd
}
