package command

import (
	"context"
	"github.com/spf13/cobra"
)

type launchOptions struct {
}

func NewLaunchCommand() *cobra.Command {
	var opts launchOptions

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start GHA runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLaunch(cmd.Context(), &opts)
		},
	}

	return cmd
}

func runLaunch(context context.Context, r *launchOptions) error {
	return nil
}
