/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package register

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
	"connectrpc.com/connect"
	coreconfig "drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
	giteaconfig "drassi.run/gitea-runner/config"
	"drassi.run/gitea-runner/pkg/gitea"
	pingv1 "gitea.dev/actionslib/ping/v1"
	runnerv1 "gitea.dev/actionslib/runner/v1"
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

func (r *register) Run(ctx context.Context) error {
	if r.url == "" {
		inquiry := huh.NewInput().
			Title("Gitea instance URL?").
			Value(&r.url).
			Placeholder("https://gitea.com").
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Gitea instance URL: %s\n", r.url)
	}
	if strings.HasPrefix(r.url, "https://") {
		inquiry := huh.NewConfirm().
			Title("Skip verify server TLS").
			Value(&r.insecureSkipTLSVerify)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Skip verify server TLS: %s\n", r.url)
	}
	if r.token == "" {
		inquiry := huh.NewInput().
			Title("Runner registration token?").
			Value(&r.token).
			EchoMode(huh.EchoModePassword).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Runner registration token: %s\n", r.token)
	}
	if r.name == "" {
		inquiry := huh.NewInput().
			Title("Runner name?").
			Value(&r.name).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		}

		fmt.Printf("Runner name: %s\n", r.name)
	}

	// TODO: prompt
	r.labels = []string{"ubuntu-latest", "ubuntu-22.04"}
	if err := r.selectSandboxer(ctx); err != nil {
		return err
	}

	if runner, err := r.doRegister(ctx); err != nil {
		return err
	} else {
		return r.saveConfig(runner)
	}
}

func (r *register) selectSandboxer(_ context.Context) error {
	providers := sandboxer.SupportedProviders()
	slices.Sort(providers)

	if len(providers) == 0 {
		return fmt.Errorf("no sandboxer available")
	}

	if s := r.sandboxer; s != "" {
		if slices.Contains(providers, s) {
			return fmt.Errorf("unknown sandboxer %q", s)
		}
		return nil
	}

	o := make([]huh.Option[string], 0, len(providers))
	for _, p := range providers {
		o = append(o, huh.NewOption(p, p))
	}

	// set default choice
	if slices.Contains(providers, "host") {
		r.sandboxer = "host"
	}

	inquiry := huh.NewSelect[string]().
		Title("Select the sandboxer?").
		Options(o...).
		Value(&r.sandboxer)

	if err := inquiry.Run(); err != nil {
		return err
	}

	fmt.Printf("Sandboxer: %s\n", r.sandboxer)
	return nil
}

func (r *register) doRegister(ctx context.Context) (*giteaconfig.Runner, error) {
	client := gitea.NewClient(r.url, r.insecureSkipTLSVerify, "", "")

	for {
		req := connect.NewRequest(&pingv1.PingRequest{
			Data: r.name,
		})
		if _, err := client.Ping(ctx, req); err == nil {
			break
		}
	}

	resp, err := client.Register(ctx, connect.NewRequest(&runnerv1.RegisterRequest{
		Name:    r.name,
		Token:   r.token,
		Version: "dev",
		Labels:  r.labels,
	}))
	if err != nil {
		fmt.Printf("cannot register new runner")
		return nil, err
	}

	runner := giteaconfig.Runner{
		Name:                  resp.Msg.Runner.Name,
		UUID:                  resp.Msg.Runner.Uuid,
		Token:                 resp.Msg.Runner.Token,
		Address:               r.url,
		InsecureSkipTLSVerify: r.insecureSkipTLSVerify,
		RunnerLabels:          resp.Msg.Runner.Labels,
	}

	return &runner, nil
}

func (r *register) saveConfig(runner *giteaconfig.Runner) error {
	config := giteaconfig.DefaultConfig()
	config.Runner = runner
	if sbConfig, err := r.defaultSandboxerConfig(r.sandboxer); err != nil {
		return err
	} else {
		config.Sandboxer = &coreconfig.Sandboxer{
			Provider: r.sandboxer,
			Config:   sbConfig,
		}
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

func (r *register) defaultSandboxerConfig(provider string) ([]byte, error) {
	cfg := sandboxer.DefaultConfig(provider)
	if cfg == nil {
		return nil, nil
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	var a any
	if err = json.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	return toml.Marshal(a)
}

// IsNotEmpty requires a non-empty string.
func IsNotEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("required value")
	}

	return nil
}
