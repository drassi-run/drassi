// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package clientcredentials implements the OAuth2.0 "client credentials" token flow,
// also known as the "two-legged OAuth 2.0".
//
// This should be used when the client is acting on its own behalf or when the client
// is the resource owner. It may also be used when requesting access to protected
// resources based on an authorization previously arranged with the authorization
// server.
//
// See https://tools.ietf.org/html/rfc6749#section-4.4
// See https://tools.ietf.org/html/rfc7523
//
// Cloned from https://github.com/golang/oauth2/pull/450 with some refactor
package clientcredentials

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dungdm93/drassi/core/pkg/util/oauth2/internal"
	"golang.org/x/oauth2"
)

type AuthMethod int

const (
	AuthMethodAutoDetect        AuthMethod = 0
	AuthMethodClientSecretBasic AuthMethod = 1
	AuthMethodClientSecretPost  AuthMethod = 2
	AuthMethodClientSecretJwt   AuthMethod = 3
	AuthMethodPrivateKeyJwt     AuthMethod = 4
)

// Config describes a 2-legged OAuth2 flow, with both the
// client application information and the server's endpoint URLs.
type Config struct {
	// ClientID is the application's ID.
	ClientID string

	// ClientSecret is the application's secret.
	ClientSecret string

	// TokenURL is the resource server's token endpoint
	// URL. This is a constant specific to each server.
	TokenURL string

	// Scope specifies optional requested permissions.
	Scopes []string

	// EndpointParams specifies additional parameters for requests to the token endpoint.
	EndpointParams url.Values

	// AuthMethod optionally specifies how the client authenticate to the TokenURL.
	// The zero value means to auto-detect.
	// See https://openid.net/specs/openid-connect-core-1_0.html#ClientAuthentication.
	AuthMethod AuthMethod

	// authStyleCache caches which auth style to use when Endpoint.AuthStyle is
	// the zero value (AuthStyleAutoDetect).
	authStyleCache internal.LazyAuthStyleCache

	// JWTExpires optionally specifies how long the jwt token is valid for.
	JWTExpires time.Duration

	// PrivateKey is RSA private key used to sign JWT assertion.
	// It's required when AuthMethod = AuthMethodPrivateKeyJwt
	PrivateKey *rsa.PrivateKey

	// KeyID contains an optional hint indicating which key is being used.
	KeyID string

	OnTokenRetrieved func(*oauth2.Token) error
}

// Token uses client credentials to retrieve a token.
//
// The provided context optionally controls which HTTP client is used. See the oauth2.HTTPClient variable.
func (c *Config) Token(ctx context.Context) (*oauth2.Token, error) {
	return c.TokenSource(ctx).Token()
}

// Client returns an HTTP client using the provided token.
// The token will auto-refresh as necessary.
//
// The provided context optionally controls which HTTP client
// is returned. See the oauth2.HTTPClient variable.
//
// The returned Client and its Transport should not be modified.
func (c *Config) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx))
}

// TokenSource returns a TokenSource that returns t until t expires,
// automatically refreshing it as necessary using the provided context and the
// client ID and client secret.
//
// Most users will use Config.Client instead.
func (c *Config) TokenSource(ctx context.Context) oauth2.TokenSource {
	source := &tokenSource{
		ctx:  ctx,
		conf: c,
	}
	return oauth2.ReuseTokenSource(nil, source)
}

type tokenSource struct {
	ctx  context.Context
	conf *Config
}

// Token refreshes the token by using a new client credentials request.
// tokens received this way do not include a refresh token
func (c *tokenSource) Token() (*oauth2.Token, error) {
	v := url.Values{
		"grant_type": {"client_credentials"},
	}
	if len(c.conf.Scopes) > 0 {
		v.Set("scope", strings.Join(c.conf.Scopes, " "))
	}
	for k, p := range c.conf.EndpointParams {
		// Allow grant_type to be overridden to allow interoperability with
		// non-compliant implementations.
		if _, ok := v[k]; ok && k != "grant_type" {
			return nil, fmt.Errorf("oauth2: cannot overwrite parameter %q", k)
		}
		v[k] = p
	}

	var authStyle internal.AuthStyle

	switch c.conf.AuthMethod {
	case AuthMethodAutoDetect:
		authStyle = internal.AuthStyleUnknown
	case AuthMethodClientSecretBasic:
		authStyle = internal.AuthStyleInHeader
	case AuthMethodClientSecretPost:
		authStyle = internal.AuthStyleInParams
	case AuthMethodClientSecretJwt:
		return nil, fmt.Errorf("client_secret_jwt AuthMethod is not supported yet")
	case AuthMethodPrivateKeyJwt:
		authStyle = internal.AuthStyleInParams
		v.Set("client_assertion_type", clientAssertionType)

		if assertion, err := c.jwtAssertion(); err != nil {
			return nil, err
		} else {
			v.Set("client_assertion", assertion)
		}
	}

	token, err := internal.RetrieveToken(c.ctx, c.conf.ClientID, c.conf.ClientSecret, c.conf.TokenURL, v, authStyle, c.conf.authStyleCache.Get())
	if err != nil {
		return nil, err
	}

	if c.conf.OnTokenRetrieved != nil {
		if err := c.conf.OnTokenRetrieved(token); err != nil {
			return nil, err
		}
	}
	return token, nil
}
