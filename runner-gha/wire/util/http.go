package util

import (
	"context"
	"net/http"

	"drassi.run/gha-runner/pkg/messages"
	"golang.org/x/oauth2"
)

func Oauth2Client(ep *messages.ServiceEndpoint) (*http.Client, error) {
	hc := http.DefaultClient
	if source, err := ep.TokenSource(); err != nil {
		return nil, err
	} else if source != nil {
		hc = oauth2.NewClient(context.Background(), source)
	}
	return hc, nil
}
