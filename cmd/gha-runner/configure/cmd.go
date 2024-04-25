package configure

import (
	"github.com/spf13/cobra"
)

func NewConfigureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"config"},
		Short:   "Configure GitHub Actions Runner",
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(
		NewRegisterCommand(),
	)

	return cmd
}
