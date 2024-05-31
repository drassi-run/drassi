package service

import (
	"code.gitea.io/actions-proto-go/ping/v1/pingv1connect"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
)

type Client interface {
	pingv1connect.PingServiceClient
	runnerv1connect.RunnerServiceClient
	Address() string
	Insecure() bool
}

var _ Client = (*httpClient)(nil)
