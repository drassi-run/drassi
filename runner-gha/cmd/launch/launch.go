/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package launch

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"time"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/oauth2/clientcredentials"
	ghav1a1 "drassi.run/gha-runner/pkg/apis/v1alpha1"
	"drassi.run/gha-runner/pkg/listener"
	"drassi.run/gha-runner/pkg/message"
	"github.com/chainguard-dev/clog"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
)

type options struct {
	Store     string
	ConfigDir string
	Name      string
}

type launcher struct {
	options

	Runner      *ghav1a1.GitHubRunner
	Key         *rsa.PrivateKey
	Sandboxer   sandboxer.Engine
	TokenSource oauth2.TokenSource
}

func New() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start GHA runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			l := launcher{options: opts}

			if err := l.Init(ctx); err != nil {
				return err
			}
			return l.Run(ctx)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Store, "store", "local", "Manifest store")

	flags.StringVar(&opts.ConfigDir, "config-dir", "", "Configuration directory")
	_ = cobra.MarkFlagDirname(flags, "config-dir")

	flags.StringVar(&opts.Name, "name", "", "GHA Runner instance name")
	_ = cobra.MarkFlagRequired(flags, "name")

	return cmd
}

func (l *launcher) Init(ctx context.Context) (err error) {
	clog.InfoContextf(ctx, "initializing gha-runner")

	store, err := manifestStore(&l.options)
	if err != nil {
		return err
	}

	if runner, err := loadRunnerManifest(ctx, store, l.Name); err != nil {
		return err
	} else {
		l.Runner = runner
	}

	spec := l.Runner.Spec
	secretName := spec.Authorization.SecretRef.Name
	if secret, err := loadSecretManifest(ctx, store, secretName); err != nil {
		return err
	} else if key, err := decodeKey(secret); err != nil {
		return err
	} else {
		l.Key = key
	}

	if sb, err := loadSandboxer(ctx, store, spec.SandboxerRef); err != nil {
		return err
	} else {
		l.Sandboxer = sb
	}

	authz := spec.Authorization
	config := clientcredentials.Config{
		TokenURL:   authz.Url,
		ClientID:   authz.ClientId,
		AuthMethod: clientcredentials.AuthMethodPrivateKeyJwt,
		PrivateKey: l.Key,

		// GHA allowed maximum lifetime 5min
		// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/Jwt/JsonWebToken.cs#L41
		JWTExpires:       5 * time.Minute,
		OnTokenRetrieved: fixupToken,
	}
	l.TokenSource = config.TokenSource(ctx)

	return nil
}

func (l *launcher) Run(ctx context.Context) error {
	clog.InfoContextf(ctx, "gha-runner started")
	hc := oauth2.NewClient(ctx, l.TokenSource)

	spec := l.Runner.Spec
	lis, err := listener.NewListener(spec.ServerUrl, hc)
	if err != nil {
		return err
	}

	if err = lis.CreateSession(ctx, spec.RunnerId, spec.GroupId, l.Key); err != nil {
		return err
	}
	defer lis.DeleteSession(ctx)

	// fetchInterval = 1s
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	for {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}

		if msg, err := lis.GetMessage(ctx, "Linux", "X64"); err != nil {
			return err
		} else if err = l.handleMessage(ctx, msg); err != nil {
			return err
		}
	}
}

func (l *launcher) handleMessage(ctx context.Context, msg *message.Message) error {
	if msg == nil || msg.Type == "" {
		return nil
	}

	switch msg.Type {
	case message.TypeAgentRefresh:
		if msg, err := message.Decode[message.AgentRefresh](msg.Body); err != nil {
			return err
		} else {
			return l.refreshAgent(ctx, msg)
		}
	case message.TypeRunnerRefresh:
		if msg, err := message.Decode[message.RunnerRefresh](msg.Body); err != nil {
			return err
		} else {
			return l.refreshRunner(ctx, msg)
		}
	case message.TypeRunnerShutdown:
		if msg, err := message.Decode[message.RunnerShutdown](msg.Body); err != nil {
			return err
		} else {
			return l.shutdownRunner(ctx, msg)
		}
	case message.TypeJobCancelMessage:
		if msg, err := message.Decode[message.JobCancel](msg.Body); err != nil {
			return err
		} else {
			return l.cancelJob(ctx, msg)
		}
	case message.TypeRunnerJobRequest:
		if msg, err := message.Decode[message.RunnerJobRequest](msg.Body); err != nil {
			return err
		} else {
			return l.requestRunnerJob(ctx, msg)
		}
	case message.TypePipelineAgentJobRequest:
		if msg, err := message.Decode[message.PipelineAgentJobRequest](msg.Body); err != nil {
			return err
		} else {
			return l.requestPipelineAgentJob(ctx, msg)
		}
	case message.TypeForceTokenRefresh:
		return l.forceRefreshToken(ctx)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

func (l *launcher) refreshAgent(ctx context.Context, msg *message.AgentRefresh) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) refreshRunner(ctx context.Context, msg *message.RunnerRefresh) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) shutdownRunner(ctx context.Context, msg *message.RunnerShutdown) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) cancelJob(ctx context.Context, msg *message.JobCancel) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) requestRunnerJob(ctx context.Context, msg *message.RunnerJobRequest) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) requestPipelineAgentJob(ctx context.Context, msg *message.PipelineAgentJobRequest) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) forceRefreshToken(ctx context.Context) error {
	log.Printf("force refresh token")
	return nil
}
