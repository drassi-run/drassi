/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"drassi.run/core/util/oauth2/clientcredentials"
	"drassi.run/gha-runner/pkg/gha"
	"drassi.run/gha-runner/pkg/message"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type launchOptions struct {
}

type launchCommand struct {
	opts *launchOptions

	runner  *gha.Runner
	client  *gha.Client
	session *gha.Session
	key     *rsa.PrivateKey
	eKey    []byte
}

func NewLaunchCommand() *cobra.Command {
	var opts launchOptions

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start GHA runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			command := launchCommand{opts: &opts}

			defer command.finalize(ctx)
			if err := command.initialize(ctx); err != nil {
				return err
			}
			return command.run(ctx)
		},
	}

	return cmd
}

func (c *launchCommand) initialize(ctx context.Context) (err error) {
	authn := new(actionsAuth)
	if err = loadJson(".credentials", authn); err != nil {
		return err
	}
	if c.key, err = loadRSA("rsa"); err != nil {
		return err
	}

	c.runner = new(gha.Runner)
	if err = loadJson(".runner", c.runner); err != nil {
		return err
	}

	authz := c.runner.Authorization
	config := clientcredentials.Config{
		ClientID:         authz.ClientId,
		TokenURL:         authz.AuthorizationUrl,
		AuthMethod:       clientcredentials.AuthMethodPrivateKeyJwt,
		JWTExpires:       5 * time.Minute,
		PrivateKey:       c.key,
		OnTokenRetrieved: fixupToken,
	}
	if c.client, err = gha.NewClient(ctx, authn.TenantUrl, config.TokenSource(ctx)); err != nil {
		return err
	}

	var sessionName string
	if sessionName, err = os.Hostname(); err != nil {
		sessionName = "RUNNER"
	}
	session := &gha.Session{
		OwnerName: sessionName,
		Runner:    &c.runner.RunnerReference,
	}
	if c.session, err = c.client.CreateSession(ctx, 1, session); err != nil {
		return err
	}
	if c.eKey, err = c.session.GetEncryptionKey(c.key); err != nil {
		return err
	}
	return nil
}

func (c *launchCommand) finalize(ctx context.Context) {
	if c.client != nil {
		return
	}

	if c.session != nil {
		log.Printf("delete session %s", c.session.Id)
		if err := c.client.DeleteSession(ctx, 1, c.session.Id); err != nil {
			log.Printf("failed to delete session: %v", err)
		}
		c.session = nil
	}

	c.client = nil
}

func (c *launchCommand) run(ctx context.Context) error {
	opts := gha.GetMessageOptions{
		SessionId:     c.session.Id,
		RunnerVersion: "2.316.1",
		OS:            "Linux",
		Architecture:  "X64",
		DisableUpdate: true,
	}
	for {
		msg, err := c.client.GetMessage(ctx, 1, opts)
		if err != nil {
			return err
		}
		err = c.handleMessage(ctx, msg)
		if err != nil {
			return err
		}
	}
}

func (c *launchCommand) handleMessage(ctx context.Context, msg *message.Message) error {
	if msg == nil || msg.Type == "" {
		return nil
	}
	body, err := msg.DecryptBody(c.eKey)
	if err != nil {
		return err
	}

	switch msg.Type {
	case message.TypeAgentRefresh:
		if msg, err := message.Decode[message.AgentRefresh](body); err != nil {
			return err
		} else {
			return c.refreshAgent(ctx, msg)
		}
	case message.TypeRunnerRefresh:
		if msg, err := message.Decode[message.RunnerRefresh](body); err != nil {
			return err
		} else {
			return c.refreshRunner(ctx, msg)
		}
	case message.TypeRunnerShutdown:
		if msg, err := message.Decode[message.RunnerShutdown](body); err != nil {
			return err
		} else {
			return c.shutdownRunner(ctx, msg)
		}
	case message.TypeJobCancelMessage:
		if msg, err := message.Decode[message.JobCancel](body); err != nil {
			return err
		} else {
			return c.cancelJob(ctx, msg)
		}
	case message.TypeRunnerJobRequest:
		if msg, err := message.Decode[message.RunnerJobRequest](body); err != nil {
			return err
		} else {
			return c.requestRunnerJob(ctx, msg)
		}
	case message.TypePipelineAgentJobRequest:
		if msg, err := message.Decode[message.PipelineAgentJobRequest](body); err != nil {
			return err
		} else {
			return c.requestPipelineAgentJob(ctx, msg)
		}
	case message.TypeForceTokenRefresh:
		return c.forceRefreshToken(ctx)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

func (c *launchCommand) refreshAgent(ctx context.Context, msg *message.AgentRefresh) error {
	log.Printf("%#v", msg)
	return nil
}

func (c *launchCommand) refreshRunner(ctx context.Context, msg *message.RunnerRefresh) error {
	log.Printf("%#v", msg)
	return nil
}

func (c *launchCommand) shutdownRunner(ctx context.Context, msg *message.RunnerShutdown) error {
	log.Printf("%#v", msg)
	return nil
}

func (c *launchCommand) cancelJob(ctx context.Context, msg *message.JobCancel) error {
	log.Printf("%#v", msg)
	return nil
}

func (c *launchCommand) requestRunnerJob(ctx context.Context, msg *message.RunnerJobRequest) error {
	log.Printf("%#v", msg)
	return nil
}

func (c *launchCommand) requestPipelineAgentJob(ctx context.Context, msg *message.PipelineAgentJobRequest) error {
	log.Printf("%#v", msg)
	return nil
}

func (c *launchCommand) forceRefreshToken(ctx context.Context) error {
	log.Printf("force refresh token")
	return nil
}

func fixupToken(token *oauth2.Token) error {
	if token != nil && strings.EqualFold(token.TokenType, "jwt") {
		token.TokenType = "Bearer"
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

func loadRSA(file string) (*rsa.PrivateKey, error) {
	if data, err := os.ReadFile(file); err != nil {
		return nil, err
	} else {
		block, _ := pem.Decode(data)
		data = block.Bytes
		key, err := x509.ParsePKCS8PrivateKey(data)
		if err != nil {
			key, err = x509.ParsePKCS1PrivateKey(data)
			if err != nil {
				return nil, fmt.Errorf("private key should be a PEM or plain PKCS1 or PKCS8; parse error: %v", err)
			}
		}
		if k, ok := key.(*rsa.PrivateKey); !ok {
			return nil, fmt.Errorf("private key is invalid")
		} else {
			return k, nil
		}
	}
}
