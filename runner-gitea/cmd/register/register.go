/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package register

import (
	"context"
	"fmt"
	"os"
	"strings"

	pingv1 "code.gitea.io/actions-proto-go/ping/v1"
	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"connectrpc.com/connect"
	giteaconfig "drassi.run/gitea-runner/config"
	"drassi.run/gitea-runner/pkg/gitea"
	"github.com/charmbracelet/huh"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

type options struct {
	url                   string
	token                 string
	name                  string
	labels                []string
	insecureSkipTLSVerify bool
	sandboxer             string
}

type register struct {
	options
}

func New() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new runner to the Gitea server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			command := register{options: opts}

			return command.Run(ctx)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.url, "url", "", "Gitea instance URL")
	flags.BoolVar(&opts.insecureSkipTLSVerify, "insecure-skip-tls-verify", false, "Skip verification of server certificate")
	flags.StringVar(&opts.token, "token", "", "Runner registration token")
	flags.StringVar(&opts.name, "name", "", "Runner name")
	flags.StringSliceVar(&opts.labels, "labels", nil, "Runner tags, comma separated")
	flags.StringVar(&opts.sandboxer, "sandboxer", "", "Sandboxer name to used")

	return cmd
}

func (c *register) Run(ctx context.Context) error {
	if c.url == "" {
		inquiry := huh.NewInput().
			Title("Gitea instance URL?").
			Value(&c.url).
			Placeholder("https://gitea.com").
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Gitea instance URL: %s\n", c.url)
	}
	if strings.HasPrefix(c.url, "https://") {
		inquiry := huh.NewConfirm().
			Title("Skip verify server TLS").
			Value(&c.insecureSkipTLSVerify)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Skip verify server TLS: %s\n", c.url)
	}
	if c.token == "" {
		inquiry := huh.NewInput().
			Title("Runner registration token?").
			Value(&c.token).
			EchoMode(huh.EchoModePassword).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Runner registration token: %s\n", c.token)
	}
	if c.name == "" {
		inquiry := huh.NewInput().
			Title("Runner name?").
			Value(&c.name).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Runner name: %s\n", c.name)
	}

	// TODO: prompt
	c.labels = []string{"ubuntu-latest", "ubuntu-22.04"}

	if c.sandboxer == "" {
		inquiry := huh.NewInput().
			Title("Sandboxer Name?").
			Value(&c.sandboxer).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Sandboxer Name: %s\n", c.sandboxer)
	}

	if runner, err := c.doRegister(ctx); err != nil {
		return err
	} else {
		return c.saveConfig(runner)
	}
}

func (c *register) doRegister(ctx context.Context) (*giteaconfig.Runner, error) {
	client := gitea.NewClient(c.url, c.insecureSkipTLSVerify, "", "")

	for {
		req := connect.NewRequest(&pingv1.PingRequest{
			Data: c.name,
		})
		if _, err := client.Ping(ctx, req); err == nil {
			break
		}
	}

	resp, err := client.Register(ctx, connect.NewRequest(&runnerv1.RegisterRequest{
		Name:    c.name,
		Token:   c.token,
		Version: "dev",
		Labels:  c.labels,
	}))
	if err != nil {
		fmt.Printf("cannot register new runner")
		return nil, err
	}

	runner := giteaconfig.Runner{
		Name:                  resp.Msg.Runner.Name,
		UUID:                  resp.Msg.Runner.Uuid,
		Token:                 resp.Msg.Runner.Token,
		Address:               c.url,
		InsecureSkipTLSVerify: c.insecureSkipTLSVerify,
		RunnerLabels:          resp.Msg.Runner.Labels,
	}

	return &runner, nil
}

func (c *register) saveConfig(runner *giteaconfig.Runner) error {
	config := &giteaconfig.Config{
		Runner:       runner,
		UseSandboxer: c.sandboxer,
	}
	b, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, strings.Repeat("=", 50))
	if _, err = os.Stdout.Write(b); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, strings.Repeat("=", 50))

	saveToFile := false
	confirm := huh.NewConfirm().
		Title("Do you want to save it to file?").
		Value(&saveToFile)
	if err = confirm.Run(); err != nil {
		return err
	} else if !saveToFile {
		return nil
	}

	var f string
	inquiry := huh.NewInput().
		Title("Select file").
		Value(&f).
		Validate(IsNotEmpty)
	if err = inquiry.Run(); err != nil {
		return err
	}

	file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	return err
}

// IsNotEmpty requires a non-empty string.
func IsNotEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("required value")
	}

	return nil
}
