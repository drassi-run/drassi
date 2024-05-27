package cmd

import (
	"github.com/spf13/cobra"
)

func NewGiteaRunnerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitea-runner",
		Short: "Gitea runner",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		NewRegisterCommand(),
	)

	return cmd
}
