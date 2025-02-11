package cmd

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
	"drassi.run/gitea-runner/pkg/service"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type registerOptions struct {
	url                   string
	token                 string
	name                  string
	labels                []string
	insecureSkipTLSVerify bool
	sandboxerKind         string
	sandboxerName         string
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
	flags.StringVar(&opts.sandboxerKind, "sandboxer-kind", "", "Sandboxer kind")
	flags.StringVar(&opts.sandboxerName, "sandboxer-name", "", "Sandboxer name")

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

	if c.opts.sandboxerKind == "" {
		err := huh.NewInput().
			Title("Sandboxer Kind?").
			Value(&c.opts.sandboxerKind).
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Sandboxer Kind: %s\n", c.opts.sandboxerKind)
		}
	}

	if c.opts.sandboxerName == "" {
		err := huh.NewInput().
			Title("Sandboxer Name?").
			Value(&c.opts.sandboxerName).
			Validate(IsNotEmpty).
			Run()
		if err != nil {
			return err
		} else {
			fmt.Printf("Sandboxer Name: %s\n", c.opts.sandboxerName)
		}
	}

	if runner, err := c.doRegister(ctx); err != nil {
		return err
	} else {
		return c.saveManifest(runner)
	}
}

func (c *registerCommand) doRegister(ctx context.Context) (*giteav1a1.GiteaRunner, error) {
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
		return nil, err
	}

	apiGroup := sandboxerv1a1.SchemeGroupVersion.String()
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
			Address:               c.opts.url,
			InsecureSkipTLSVerify: c.opts.insecureSkipTLSVerify,
			RunnerLabels:          resp.Msg.Runner.Labels,
			SandboxerRef: corev1.TypedLocalObjectReference{
				APIGroup: &apiGroup,
				Kind:     c.opts.sandboxerKind,
				Name:     c.opts.sandboxerName,
			},
		},
	}

	return &runner, nil
}

func (c *registerCommand) saveManifest(runner *giteav1a1.GiteaRunner) error {
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
	err = huh.NewConfirm().
		Title("Do you want to save it to file?").
		Value(&saveToFile).
		Run()
	if err != nil {
		return err
	} else if !saveToFile {
		return nil
	}

	var f string
	err = huh.NewInput().
		Title("Select file").
		Value(&f).
		Validate(IsNotEmpty).
		Run()
	if err != nil {
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
