/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitea

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"code.gitea.io/actions-proto-go/ping/v1/pingv1connect"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
	"connectrpc.com/connect"
)

type Client interface {
	pingv1connect.PingServiceClient
	runnerv1connect.RunnerServiceClient
	Address() string
	Insecure() bool
}

var _ Client = (*client)(nil)

// A client manages communication with the runner API.
type client struct {
	pingv1connect.PingServiceClient
	runnerv1connect.RunnerServiceClient
	endpoint string
	insecure bool
}

func (c *client) Address() string {
	return c.endpoint
}

func (c *client) Insecure() bool {
	return c.insecure
}

// NewClient returns a new Gitea runner client.
func NewClient(endpoint string, insecure bool, uuid, token string, opts ...connect.ClientOption) Client {
	baseURL := strings.TrimRight(endpoint, "/") + "/api/actions"

	var interceptor connect.UnaryInterceptorFunc = func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if uuid != "" {
				req.Header().Set(UUIDHeader, uuid)
			}
			if token != "" {
				req.Header().Set(TokenHeader, token)
			}
			return next(ctx, req)
		}
	}
	opts = append(opts, connect.WithInterceptors(interceptor))
	hc := getHTTPClient(endpoint, insecure)

	return &client{
		PingServiceClient:   pingv1connect.NewPingServiceClient(hc, baseURL, opts...),
		RunnerServiceClient: runnerv1connect.NewRunnerServiceClient(hc, baseURL, opts...),
		endpoint:            endpoint,
		insecure:            insecure,
	}
}

func getHTTPClient(endpoint string, insecure bool) *http.Client {
	if strings.HasPrefix(endpoint, "https://") && insecure {
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}
	return http.DefaultClient
}
