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
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dungdm93/drassi/core/pkg/util/oauth2/clientcredentials"
	"github.com/dungdm93/drassi/gha-runner/pkg/gha"
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
			command := launchCommand{opts: &opts}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			go func() {
				<-ctx.Done()
				command.finalize(ctx)
			}()

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
		err = c.handleMessage(msg)
		if err != nil {
			return err
		}
	}
}

func (c *launchCommand) handleMessage(msg *gha.Message) error {
	if msg == nil || msg.Type == "" {
		return nil
	}
	body, err := msg.DecryptBody(c.eKey)
	if err != nil {
		return err
	}

	switch msg.Type {
	case gha.MessageTypeAgentRefresh:
		message := new(gha.AgentRefreshMessage)
		if err := json.Unmarshal(body, message); err != nil {
			return err
		}
		return c.refreshAgent(message)
	case gha.MessageTypeRunnerRefresh:
		message := new(gha.RunnerRefreshMessage)
		if err := json.Unmarshal(body, message); err != nil {
			return err
		}
		return c.refreshRunner(message)
	case gha.MessageTypeRunnerShutdown:
		message := new(gha.RunnerShutdownMessage)
		if err := json.Unmarshal(body, message); err != nil {
			return err
		}
		return c.shutdownRunner(message)
	case gha.MessageTypeJobCancelMessage:
		message := new(gha.JobCancelMessage)
		if err := json.Unmarshal(body, message); err != nil {
			return err
		}
		return c.cancelJob(message)
	case gha.MessageTypeRunnerJobRequest:
		message := new(gha.RunnerJobRequestMessage)
		if err := json.Unmarshal(body, message); err != nil {
			return err
		}
		return c.requestRunnerJob(message)
	case gha.MessageTypePipelineAgentJobRequest:
		message := new(gha.PipelineAgentJobRequestMessage)
		if err := json.Unmarshal(body, message); err != nil {
			return err
		}
		return c.requestPipelineAgentJob(message)
	case gha.MessageTypeForceTokenRefresh:
		return c.forceRefreshToken()
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

func (c *launchCommand) refreshAgent(message *gha.AgentRefreshMessage) error {
	log.Printf("%#v", message)
	return nil
}

func (c *launchCommand) refreshRunner(message *gha.RunnerRefreshMessage) error {
	log.Printf("%#v", message)
	return nil
}

func (c *launchCommand) shutdownRunner(message *gha.RunnerShutdownMessage) error {
	log.Printf("%#v", message)
	return nil
}

func (c *launchCommand) cancelJob(message *gha.JobCancelMessage) error {
	log.Printf("%#v", message)
	return nil
}

func (c *launchCommand) requestRunnerJob(message *gha.RunnerJobRequestMessage) error {
	log.Printf("%#v", message)
	return nil
}

func (c *launchCommand) requestPipelineAgentJob(message *gha.PipelineAgentJobRequestMessage) error {
	log.Printf("%#v", message)
	return nil
}

func (c *launchCommand) forceRefreshToken() error {
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
