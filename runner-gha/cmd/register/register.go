/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package register

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"strings"

	sandboxerv1a1 "drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	"drassi.run/core/util/http"
	ghav1a1 "drassi.run/gha-runner/pkg/apis/v1alpha1"
	"drassi.run/gha-runner/pkg/dotnet"
	"drassi.run/gha-runner/pkg/types"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	groupEndpoint  = "_apis/distributedtask/pools"
	runnerEndpoint = groupEndpoint + "/%d/agents"
)

type actionsAuth struct {
	TenantUrl   string `json:"url"`
	TokenSchema string `json:"token_schema"`
	Token       string `json:"token"`
	UseV2Flow   bool   `json:"use_v2_flow"`
}

type options struct {
	Url           string
	Token         string
	Group         string
	Name          string
	SandboxerKind string
	SandboxerName string
}

var UserAgent types.UserAgentInfo

type register struct {
	options

	client *xhttp.Client
	auth   *actionsAuth
	key    *rsa.PrivateKey
	group  *types.Group
	runner *types.Runner
}

func New() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register new runner to the GitHub Actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			r := &register{options: opts}
			return r.Run(ctx)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Url, "url", "", "GitHub Actions URL")
	flags.StringVar(&opts.Token, "token", "", "Actions registration Token")
	flags.StringVar(&opts.Group, "group", "", "GitHub Actions RunnerReference Group")
	flags.StringVar(&opts.Name, "name", "", "GitHub Actions RunnerReference Name")
	flags.StringVar(&opts.SandboxerKind, "sandboxer-kind", "", "Sandboxer kind")
	flags.StringVar(&opts.SandboxerName, "sandboxer-name", "", "Sandboxer name")

	return cmd
}

func (r *register) Run(ctx context.Context) error {
	// Using registration token to authenticate with GitHub API
	if err := r.authenticate(ctx); err != nil {
		return err
	}

	// Select Runner group
	if err := r.selectRunnerGroup(ctx); err != nil {
		return err
	}

	// Provide Runner name
	if err := r.provideRunnerName(ctx); err != nil {
		return err
	}

	// Select Sandboxer
	if err := r.selectSandboxer(ctx); err != nil {
		return err
	}

	// Generate RSA key
	if key, err := rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return err
	} else {
		r.key = key
	}

	// Register runner to GitHub API
	if err := r.registerRunner(ctx); err != nil {
		return err
	}

	// Save runner configuration
	return r.saveRunner(ctx)
}

func (r *register) getApiUrl() (string, error) {
	u, err := url.Parse(r.Url)
	if err != nil {
		return "", err
	}

	scheme := strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Host)
	if host == "github.com" || host == "www.github.com" || host == "github.localhost" ||
		strings.HasSuffix(host, ".ghe.com") ||
		strings.HasSuffix(host, ".ghe.localhost") {
		apiUrl := url.URL{
			Scheme: scheme,
			Host:   "api." + host,
		}
		return apiUrl.String(), nil
	} else {
		apiUrl := url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   "/api/v3",
		}
		return apiUrl.String(), nil
	}
}

func (r *register) authenticate(ctx context.Context) error {
	if r.Url == "" {
		inquiry := huh.NewInput().
			Title("What is the URL of your repository?").
			Value(&r.Url).
			Placeholder("https://github.com/[ORG]/[REPO]").
			Validate(IsNotEmpty)

		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("GitHub Actions URL: %s\n", r.Url)
		}
	}
	if r.Token == "" {
		inquiry := huh.NewInput().
			Title("What is your runner register Token?").
			EchoMode(huh.EchoModePassword).
			Value(&r.Token).
			Validate(IsNotEmpty)

		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("Actions registration Token: %s\n", r.Token)
		}
	}

	var client *xhttp.Client
	if apiUrl, err := r.getApiUrl(); err != nil {
		return err
	} else if client, err = xhttp.NewClient(apiUrl); err != nil {
		return err
	}
	client = client.WithDefaultHeader("User-Agent", UserAgent.String())

	var auth actionsAuth
	data := map[string]string{
		"url":          r.Url,
		"runner_event": "register",
	}
	hr := client.Post("/actions/runner-registration").
		SetHeader("Authorization", "RemoteAuth "+r.Token).
		WithBodyProvider(xhttp.JsonEncode(data)).
		OnSuccess(xhttp.JsonDecode(&auth))

	spin := spinner.New().
		Context(ctx).
		Title("Authenticate to GitHub API").
		ActionWithErr(hr.Do)

	if err := spin.Run(); err != nil {
		return err
	} else {
		r.auth = &auth
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: auth.Token,
		TokenType:   "Bearer",
	})

	if c, err := xhttp.NewClient(auth.TenantUrl); err != nil {
		return err
	} else {
		r.client = c.
			WithHttpClient(oauth2.NewClient(ctx, tokenSource)).
			WithDefaultHeader("User-Agent", UserAgent.String())
		return nil
	}
}

func (r *register) selectRunnerGroup(ctx context.Context) error {
	groups := new(types.List[types.Group])
	hr := r.client.Get(groupEndpoint).
		OnSuccess(xhttp.JsonDecode(groups))

	spin := spinner.New().
		Context(ctx).
		Title("Loading runner groups").
		ActionWithErr(hr.Do)

	if err := spin.Run(); err != nil {
		return err
	}

	if r.Group != "" {
		for _, g := range groups.Value {
			if g.IsHosted {
				continue
			}
			if g.Name == r.Group {
				r.group = &g
				return nil
			}
		}
		return fmt.Errorf("could not find any self-hosted runner Group named %s", r.Group)
	} else {
		var selectOptions []huh.Option[types.Group]
		for _, g := range groups.Value {
			if g.IsHosted {
				continue
			}
			selectOptions = append(selectOptions, huh.NewOption(g.Name, g))
		}
		var group types.Group
		inquiry := huh.NewSelect[types.Group]().
			Title("Select the runner Group to add this runner to?").
			Options(selectOptions...).
			Value(&group)

		if err := inquiry.Run(); err != nil {
			return err
		}
		r.group = &group
		return nil
	}
}

func (r *register) checkRunnerExist(ctx context.Context, name string) error {
	runners := new(types.List[types.RunnerReference])

	hr := r.client.Get(fmt.Sprintf(runnerEndpoint, r.group.Id)).
		SetQuery("agentName", name).
		OnSuccess(xhttp.JsonDecode(runners))

	spin := spinner.New().
		Context(ctx).
		Title("Validating runner name").
		ActionWithErr(hr.Do)

	if err := spin.Run(); err != nil {
		return err
	}
	if runners.Count > 0 {
		return fmt.Errorf("a runner exists with the same Name %s", name)
	}
	return nil
}

func (r *register) provideRunnerName(ctx context.Context) error {
	if r.Name == "" {
		if hostname, err := os.Hostname(); err == nil {
			r.Name = hostname // default runner Name
		}
		inquiry := huh.NewInput().
			Title("Enter the Name of runner?").
			Value(&r.Name).
			Validate(IsNotEmpty)

		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("GitHub Actions Runner Name: %s\n", r.Name)
		}
	}
	return r.checkRunnerExist(ctx, r.Name)
}

func (r *register) selectSandboxer(_ context.Context) error {
	if r.SandboxerKind == "" {
		inquiry := huh.NewInput().
			Title("What is your sandboxer kind?").
			Value(&r.SandboxerKind).
			Validate(IsNotEmpty)

		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("SandboxerKind: %s\n", r.SandboxerKind)
		}
	}
	if r.SandboxerName == "" {
		inquiry := huh.NewInput().
			Title("What is your sandboxer name?").
			Value(&r.SandboxerName).
			Validate(IsNotEmpty)

		if err := inquiry.Run(); err != nil {
			return err
		} else {
			fmt.Printf("SandboxerName: %s\n", r.SandboxerName)
		}
	}
	return nil
}

func (r *register) registerRunner(ctx context.Context) error {
	req := &types.Runner{
		RunnerReference: types.RunnerReference{
			Name:    r.Name,
			Version: "3.0.0",
		},
		MaxParallelism: 10,
		Labels: []types.Label{
			{Name: "self-hosted", Type: types.LabelTypeSystem},
			{Name: "Linux", Type: types.LabelTypeSystem},
			{Name: "X64", Type: types.LabelTypeSystem},
		},
		Authorization: types.Authorization{
			PublicKey: dotnet.NewPublicKey(&r.key.PublicKey),
		},
	}

	runner := new(types.Runner)
	hr := r.client.Post(fmt.Sprintf(runnerEndpoint, r.group.Id)).
		SetQuery("api-version", "6.0-preview").
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(runner))

	spin := spinner.New().
		Context(ctx).
		Title("Registering new runner").
		ActionWithErr(hr.Do)

	if err := spin.Run(); err != nil {
		return err
	}
	r.runner = runner
	return nil
}

func (r *register) saveRunner(_ context.Context) error {
	secret, err := r.encodeKey()
	if err != nil {
		return err
	}

	labels := make([]string, len(r.runner.Labels))
	for i, l := range r.runner.Labels {
		labels[i] = l.Name
	}

	runner := &ghav1a1.GitHubRunner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ghav1a1.SchemeGroupVersion.String(),
			Kind:       "GitHubRunner",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              strings.ToLower(r.Name),
			CreationTimestamp: metav1.NewTime(*r.runner.CreatedOn),
		},
		Spec: ghav1a1.GitHubRunnerSpec{
			RunnerId:        r.runner.Id,
			GroupId:         r.group.Id,
			ServerUrl:       r.auth.TenantUrl,
			RegistrationUrl: r.Url,
			Authorization: ghav1a1.GitHubRunnerAuthorization{
				Url:      r.runner.Authorization.AuthorizationUrl,
				ClientId: r.runner.Authorization.ClientId,
				SecretRef: corev1.LocalObjectReference{
					Name: secret.Name,
				},
			},
			SandboxerRef: corev1.TypedLocalObjectReference{
				APIGroup: new(sandboxerv1a1.SchemeGroupVersion.String()),
				Kind:     r.SandboxerKind,
				Name:     r.SandboxerName,
			},
		},
		Status: ghav1a1.GitHubRunnerStatus{
			RunnerName: r.runner.Name,
			GroupName:  r.group.Name,
			Labels:     labels,
		},
	}

	var buf bytes.Buffer
	if b, err := yaml.Marshal(runner); err != nil {
		return err
	} else {
		buf.Write(b)
	}

	buf.WriteString("---\n")

	if b, err := yaml.Marshal(secret); err != nil {
		return err
	} else {
		buf.Write(b)
	}

	fmt.Println(strings.Repeat("=", 50))
	if _, err = os.Stdout.Write(buf.Bytes()); err != nil {
		return err
	}
	fmt.Println(strings.Repeat("=", 50))

	saveToFile := false
	var inquiry huh.Field
	inquiry = huh.NewConfirm().
		Title("Do you want to save it to file?").
		Value(&saveToFile)

	if err = inquiry.Run(); err != nil {
		return err
	} else if !saveToFile {
		return nil
	}

	var f string
	inquiry = huh.NewInput().
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

	_, err = buf.WriteTo(file)
	return err
}

func (r *register) encodeKey() (*corev1.Secret, error) {
	pub := x509.MarshalPKCS1PublicKey(&r.key.PublicKey)
	pubKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pub,
	})

	pri := x509.MarshalPKCS1PrivateKey(r.key)
	priKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pri,
	})

	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(r.Name) + "-secret",
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"public.key":  string(pubKey),
			"private.key": string(priKey),
		},
	}
	return &secret, nil
}
