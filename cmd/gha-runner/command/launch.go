package command

import (
	"context"
	"os"

	"encoding/json"
	"github.com/dungdm93/drasi/pkg/service/gha"
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

func runLaunch(ctx context.Context, r *launchOptions) error {
	auth := new(actionsAuth)
	if err := loadJson(".credentials", auth); err != nil {
		return err
	}
	runner := new(gha.Runner)
	if err := loadJson(".runner", runner); err != nil {
		return err
	}

	client, err := gha.NewClient(auth.TenantUrl, auth.Token)
	if err != nil {
		return err
	}

	var sessionName string
	if sessionName, err = os.Hostname(); err != nil {
		sessionName = "RUNNER"
	}
	session := &gha.Session{
		OwnerName: sessionName,
		Runner:    &runner.RunnerReference,
	}
	if session, err = client.CreateSession(ctx, 1, session); err != nil {
		return err
	}

	return nil
}

func loadJson(file string, object any) error {
	f, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	return json.NewDecoder(f).Decode(object)
}
