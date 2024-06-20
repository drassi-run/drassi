package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pingv1 "code.gitea.io/actions-proto-go/ping/v1"
	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"connectrpc.com/connect"
	"github.com/charmbracelet/huh"
	"github.com/dungdm93/drassi/gitea-runner/pkg/service"
	"github.com/spf13/cobra"
)

type registerOptions struct {
	url                   string
	token                 string
	name                  string
	labels                []string
	insecureSkipTLSVerify bool
}

type registerCommand struct {
	opts *registerOptions
}

func NewRegisterCommand() *cobra.Command {
	var opts registerOptions

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new runner to the Gitea server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := registerCommand{opts: &opts}
			ctx := cmd.Context()
			return command.run(ctx)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.url, "url", "", "Gitea instance URL")
	flags.BoolVar(&opts.insecureSkipTLSVerify, "insecure-skip-tls-verify", false, "Skip verification of server certificate")
	flags.StringVar(&opts.token, "token", "", "Runner registration token")
	flags.StringVar(&opts.name, "name", "", "Runner name")
	flags.StringSliceVar(&opts.labels, "labels", nil, "Runner tags, comma separated")

	return cmd
}

func (c *registerCommand) run(ctx context.Context) error {
	if c.opts.url == "" {
		err := huh.NewInput().
			Title("Gitea instance URL?").
			Value(&c.opts.url).
			Placeholder("https://gitea.com").
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Gitea instance URL: %s\n", c.opts.url)
		}
	}
	if strings.HasPrefix(c.opts.url, "https://") {
		err := huh.NewConfirm().
			Title("Skip verify server TLS").
			Value(&c.opts.insecureSkipTLSVerify).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Skip verify server TLS: %s\n", c.opts.url)
		}
	}
	if c.opts.token == "" {
		err := huh.NewInput().
			Title("Runner registration token?").
			Value(&c.opts.token).
			EchoMode(huh.EchoModePassword).
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Runner registration token: %s\n", c.opts.token)
		}
	}
	if c.opts.name == "" {
		err := huh.NewInput().
			Title("Runner name?").
			Value(&c.opts.name).
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Runner name: %s\n", c.opts.name)
		}
	}

	// TODO: prompt
	c.opts.labels = []string{"ubuntu-latest", "ubuntu-22.04"}

	return c.doRegister(ctx)
}

func (c *registerCommand) doRegister(ctx context.Context) error {
	client := service.NewClient(c.opts.url, c.opts.insecureSkipTLSVerify, "", "")

	for {
		_, err := client.Ping(ctx, connect.NewRequest(&pingv1.PingRequest{
			Data: c.opts.name,
		}))
		if err == nil {
			break
		}
	}

	resp, err := client.Register(ctx, connect.NewRequest(&runnerv1.RegisterRequest{
		Name:    c.opts.name,
		Token:   c.opts.token,
		Version: "dev",
		Labels:  c.opts.labels,
	}))
	if err != nil {
		fmt.Printf("cannot register new runner")
		return err
	}

	runner := &RunnerInfo{
		ID:                    resp.Msg.Runner.Id,
		UUID:                  resp.Msg.Runner.Uuid,
		Name:                  resp.Msg.Runner.Name,
		Token:                 resp.Msg.Runner.Token,
		Address:               c.opts.url,
		Labels:                resp.Msg.Runner.Labels,
		InsecureSkipTLSVerify: c.opts.insecureSkipTLSVerify,
	}

	return saveJson(".runner", runner)
}

// IsNotEmpty requires a non-empty string.
func IsNotEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("required value")
	}

	return nil
}

func saveJson(file string, object any) error {
	if f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		return enc.Encode(object)
	}
}

type RunnerInfo struct {
	ID                    int64    `json:"id,omitempty" yaml:"id,omitempty"`
	UUID                  string   `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Name                  string   `json:"name,omitempty" yaml:"name,omitempty"`
	Token                 string   `json:"token,omitempty" yaml:"token,omitempty"`
	Address               string   `json:"address,omitempty" yaml:"address,omitempty"`
	Labels                []string `json:"labels,omitempty" yaml:"labels,omitempty"`
	InsecureSkipTLSVerify bool     `json:"insecureSkipTLSVerify,omitempty" yaml:"insecureSkipTLSVerify,omitempty"`
}
