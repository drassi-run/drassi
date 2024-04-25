package configure

import (
	"fmt"
	"github.com/spf13/cobra"
)

type registerOptions struct {
	url   string
	token string
}

func NewRegisterCommand() *cobra.Command {
	var opts registerOptions

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new runner to the GitHub Actions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("configure called")
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.url, "url", "", "GitHub Actions URL")
	flags.StringVar(&opts.token, "token", "", "Actions registration token")

	return cmd
}
