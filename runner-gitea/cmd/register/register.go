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
	sandboxerv1a1 "drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	giteav1a1 "drassi.run/gitea-runner/pkg/apis/v1alpha1"
	"drassi.run/gitea-runner/pkg/gitea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type options struct {
	url                   string
	token                 string
	name                  string
	labels                []string
	insecureSkipTLSVerify bool
	sandboxerKind         string
	sandboxerName         string
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
	flags.StringVar(&opts.sandboxerKind, "sandboxer-kind", "", "Sandboxer kind")
	flags.StringVar(&opts.sandboxerName, "sandboxer-name", "", "Sandboxer name")

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
		} else {
			fmt.Printf("Gitea instance URL: %s\n", c.url)
		}
	}
	if strings.HasPrefix(c.url, "https://") {
		inquiry := huh.NewConfirm().
			Title("Skip verify server TLS").
			Value(&c.insecureSkipTLSVerify)
		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("Skip verify server TLS: %s\n", c.url)
		}
	}
	if c.token == "" {
		inquiry := huh.NewInput().
			Title("Runner registration token?").
			Value(&c.token).
			EchoMode(huh.EchoModePassword).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("Runner registration token: %s\n", c.token)
		}
	}
	if c.name == "" {
		inquiry := huh.NewInput().
			Title("Runner name?").
			Value(&c.name).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("Runner name: %s\n", c.name)
		}
	}

	// TODO: prompt
	c.labels = []string{"ubuntu-latest", "ubuntu-22.04"}

	if c.sandboxerKind == "" {
		inquiry := huh.NewInput().
			Title("Sandboxer Kind?").
			Value(&c.sandboxerKind).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("Sandboxer Kind: %s\n", c.sandboxerKind)
		}
	}

	if c.sandboxerName == "" {
		inquiry := huh.NewInput().
			Title("Sandboxer Name?").
			Value(&c.sandboxerName).
			Validate(IsNotEmpty)
		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("Sandboxer Name: %s\n", c.sandboxerName)
		}
	}

	if runner, err := c.doRegister(ctx); err != nil {
		return err
	} else {
		return c.saveManifest(runner)
	}
}

func (c *register) doRegister(ctx context.Context) (*giteav1a1.GiteaRunner, error) {
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

	runner := giteav1a1.GiteaRunner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: giteav1a1.SchemeGroupVersion.String(),
			Kind:       "GiteaRunner",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: resp.Msg.Runner.Name,
		},
		Spec: giteav1a1.GiteaRunnerSpec{
			UUID:                  resp.Msg.Runner.Uuid,
			Token:                 resp.Msg.Runner.Token,
			Address:               c.url,
			InsecureSkipTLSVerify: c.insecureSkipTLSVerify,
			RunnerLabels:          resp.Msg.Runner.Labels,
			SandboxerRef: corev1.TypedLocalObjectReference{
				APIGroup: new(sandboxerv1a1.SchemeGroupVersion.String()),
				Kind:     c.sandboxerKind,
				Name:     c.sandboxerName,
			},
		},
	}

	return &runner, nil
}

func (c *register) saveManifest(runner *giteav1a1.GiteaRunner) error {
	b, err := yaml.Marshal(runner)
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
