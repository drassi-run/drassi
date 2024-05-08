package command

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dungdm93/drasi/pkg/service/gha"
	"github.com/dungdm93/drasi/pkg/util/oauth2/clientcredentials"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type launchOptions struct {
}

func NewLaunchCommand() *cobra.Command {
	var opts launchOptions

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Start GHA runner to receive request from server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLaunch(cmd.Context(), &opts)
		},
	}

	return cmd
}

func runLaunch(ctx context.Context, r *launchOptions) error {
	authn := new(actionsAuth)
	if err := loadJson(".credentials", authn); err != nil {
		return err
	}
	key, err := loadRSA("rsa")
	if err != nil {
		return err
	}
	runner := new(gha.Runner)
	if err := loadJson(".runner", runner); err != nil {
		return err
	}
	authz := runner.Authorization

	config := clientcredentials.Config{
		ClientID:         authz.ClientId,
		TokenURL:         authz.AuthorizationUrl,
		AuthMethod:       clientcredentials.AuthMethodPrivateKeyJwt,
		JWTExpires:       5 * time.Minute,
		PrivateKey:       key,
		OnTokenRetrieved: fixupToken,
	}
	client, err := gha.NewClient(ctx, authn.TenantUrl, config.TokenSource(ctx))
	if err != nil {
		return err
	}

	var sessionName string
	if sessionName, err = os.Hostname(); err != nil {
		sessionName = "RUNNER"
	}
	session := &gha.Session{
		OwnerName: sessionName,
		Runner:    &runner.RunnerReference,
	}
	if session, err = client.CreateSession(ctx, 1, session); err != nil {
		return err
	}

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
