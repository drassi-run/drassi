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
	"net/http"
	"time"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/oauth2/clientcredentials"
	ghav1a1 "drassi.run/gha-runner/pkg/apis/v1alpha1"
	"drassi.run/gha-runner/pkg/holder"
	"drassi.run/gha-runner/pkg/listener"
	"drassi.run/gha-runner/pkg/messages"
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
	Runner      *ghav1a1.GitHubRunner
	Key         *rsa.PrivateKey
	Sandboxer   sandboxer.Engine
	TokenSource oauth2.TokenSource
	hc          *http.Client
}

func New() *cobra.Command {
	var opts options

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start GHA runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			l := new(launcher)

			if err := l.Init(ctx, &opts); err != nil {
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

func (l *launcher) Init(ctx context.Context, opts *options) (err error) {
	clog.InfoContextf(ctx, "initializing gha-runner")

	store, err := manifestStore(opts)
	if err != nil {
		return err
	}

	if runner, err := loadRunnerManifest(ctx, store, opts.Name); err != nil {
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
	l.hc = oauth2.NewClient(ctx, l.TokenSource)

	return nil
}

func (l *launcher) Run(ctx context.Context) error {
	clog.InfoContextf(ctx, "gha-runner started")

	lis, err := l.createListener()
	if err != nil {
		return err
	}

	spec := l.Runner.Spec
	cancel, err := lis.Connect(ctx, spec.RunnerId, spec.GroupId)
	if err != nil {
		return err
	}
	defer cancel()

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

func (l *launcher) createListener() (listener.Listener, error) {
	url := l.Runner.Spec.ServerUrl
	return listener.NewMigratableListener(url, l.hc, l.Key)
}

func (l *launcher) handleMessage(ctx context.Context, msg *listener.Message) error {
	if msg == nil || msg.Type == "" {
		return nil
	}

	switch msg.Type {
	case messages.TypeAgentRefresh:
		if msg, err := messages.Decode[messages.AgentRefresh](msg.Body); err != nil {
			return err
		} else {
			return l.refreshAgent(ctx, msg)
		}
	case messages.TypeRunnerRefresh:
		if msg, err := messages.Decode[messages.RunnerRefresh](msg.Body); err != nil {
			return err
		} else {
			return l.refreshRunner(ctx, msg)
		}
	case messages.TypeRunnerShutdown:
		if msg, err := messages.Decode[messages.RunnerShutdown](msg.Body); err != nil {
			return err
		} else {
			return l.shutdownRunner(ctx, msg)
		}
	case messages.TypeJobCancellation:
		if msg, err := messages.Decode[messages.JobCancel](msg.Body); err != nil {
			return err
		} else {
			return l.cancelJob(ctx, msg)
		}
	case messages.TypeRunnerJobRequest:
		if msg, err := messages.Decode[messages.RunnerJobRequest](msg.Body); err != nil {
			return err
		} else {
			return l.requestRunnerJob(ctx, msg)
		}
	case messages.TypePipelineAgentJobRequest:
		if msg, err := messages.Decode[messages.PipelineAgentJobRequest](msg.Body); err != nil {
			return err
		} else {
			return l.requestPipelineAgentJob(ctx, msg)
		}
	case messages.TypeForceTokenRefresh:
		return l.forceRefreshToken(ctx)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

func (l *launcher) refreshAgent(ctx context.Context, msg *messages.AgentRefresh) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) refreshRunner(ctx context.Context, msg *messages.RunnerRefresh) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) shutdownRunner(ctx context.Context, msg *messages.RunnerShutdown) error {
	log.Printf("%#v", msg)
	return nil
}

func (l *launcher) cancelJob(ctx context.Context, msg *messages.JobCancel) error {
	log.Printf("%#v", msg)
	return nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Listener/Runner.cs#L559-L613
func (l *launcher) requestRunnerJob(ctx context.Context, msg *messages.RunnerJobRequest) error {
	var req *messages.PipelineAgentJobRequest = nil
	if url := msg.RunServiceUrl; url != "" {
		if svc, err := holder.NewRunService(url, l.hc); err != nil {
			return err
		} else if req, err = svc.AcquireJob(ctx, msg.RunnerRequestId, msg.BillingOwnerId); err != nil {
			return err
		}
	} else {
		url = l.Runner.Spec.ServerUrl
		groupId := l.Runner.Spec.GroupId
		if svc, err := holder.NewRunnerService(url, l.hc, groupId); err != nil {
			return err
		} else if req, err = svc.AcquireJob(ctx, msg.RunnerRequestId); err != nil {
			return err
		}
	}
	return l.requestPipelineAgentJob(ctx, req)
}

func (l *launcher) requestPipelineAgentJob(ctx context.Context, msg *messages.PipelineAgentJobRequest) error {
	var lease holder.Lease

	switch msg.MessageType {
	case messages.TypeRunnerJobRequest:
		ep := msg.ServiceEndpoint("SystemVssConnection")
		if ep == nil {
			return fmt.Errorf("SystemVssConnection service endpoint not available")
		}
		if source, err := ep.TokenSource(); err != nil {
			return err
		} else {
			hc := http.DefaultClient
			if source != nil {
				hc = oauth2.NewClient(ctx, source)
			}
			if svc, err := holder.NewRunService(ep.Url, hc); err != nil {
				return err
			} else {
				lease = svc.Lease(msg)
			}
		}
	case messages.TypePipelineAgentJobRequest:
		url := l.Runner.Spec.ServerUrl
		groupId := l.Runner.Spec.GroupId
		if svc, err := holder.NewRunnerService(url, l.hc, groupId); err != nil {
			return err
		} else {
			lease = svc.Lease(msg)
		}
	default:
		return fmt.Errorf("PipelineAgentJobRequest - unsupported message type: %s", msg.MessageType)
	}

	return l.doRun(ctx, lease)
}

func (l *launcher) forceRefreshToken(ctx context.Context) error {
	log.Printf("force refresh token")
	return nil
}

func (l *launcher) doRun(ctx context.Context, lease holder.Lease) error {
	log.Printf("%#v", lease)
	return nil
}
