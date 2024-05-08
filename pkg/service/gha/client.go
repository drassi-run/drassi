package gha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	utilhttp "github.com/dungdm93/drasi/pkg/util/http"
)

const (
	groupEndpoint        = "_apis/distributedtask/pools"
	runnerEndpoint       = groupEndpoint + "/%d/agents"
	scaleSetEndpoint     = "_apis/runtime/runnerscalesets"
	apiVersionQueryParam = "api-version=6.0-preview.2"
)

type ghaResponse[T any] struct {
	Count int32 `json:"count"`
	Value []T   `json:"value"`
}

type Client struct {
	h *http.Client

	serverUrl *url.URL
	token     string
	UserAgent UserAgentInfo
}

func NewClient(serverUrl string, token string) (*Client, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return nil, err
	}

	c := Client{
		h:         &http.Client{},
		serverUrl: u,
		token:     token,
		UserAgent: UserAgentInfo{
			Version:   "1.2.3",
			CommitSHA: "abc123",
		},
	}
	return &c, nil
}

func (c *Client) send(req *http.Request) (res *http.Response, err error) {
	res, err = c.h.Do(req)
	if err != nil {
		return nil, err
	}
	if !utilhttp.IsSuccess(res.StatusCode) {
		return nil, ParseActionsErrorFromResponse(res)
	}
	return
}

func (c *Client) NewActionsServiceRequest(ctx context.Context, method, path string, query map[string]string, body io.Reader) (*http.Request, error) {
	//err := c.updateTokenIfNeeded(ctx)
	//if err != nil {
	//	return nil, err
	//}

	u := c.serverUrl.JoinPath(path)

	q := u.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	if q.Get("api-version") == "" {
		q.Set("api-version", "6.0-preview")
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("User-Agent", c.UserAgent.String())

	return req, nil
}

func (c *Client) ListGroupRunners(ctx context.Context) ([]RunnerGroup, error) {
	// Construct request
	req, err := c.NewActionsServiceRequest(ctx, http.MethodGet, groupEndpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	// Make a request
	res, err := c.send(req)
	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
	}

	// Extract the response body
	var groups ghaResponse[RunnerGroup]
	if err = json.NewDecoder(res.Body).Decode(&groups); err != nil {
		return nil, err
	}
	return groups.Value, nil
}

// groupId = 0 means all groups
func (c *Client) ListRunners(ctx context.Context, groupId int32, name string) ([]RunnerReference, error) {
	// Construct request
	query := map[string]string{
		"agentName": name,
		// includeCapabilities: false
	}
	endpoint := fmt.Sprintf(runnerEndpoint, groupId)
	req, err := c.NewActionsServiceRequest(ctx, http.MethodGet, endpoint, query, nil)
	if err != nil {
		return nil, err
	}

	// Make the request
	res, err := c.send(req)
	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
	}

	// Extract the response body
	var runners ghaResponse[RunnerReference]
	if err = json.NewDecoder(res.Body).Decode(&runners); err != nil {
		return nil, err
	}
	return runners.Value, nil
}

func (c *Client) AddRunner(ctx context.Context, groupId int32, runner *Runner) (*Runner, error) {
	// Construct request
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(runner); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(runnerEndpoint, groupId)
	req, err := c.NewActionsServiceRequest(ctx, http.MethodPost, endpoint, nil, buf)
	if err != nil {
		return nil, err
	}

	// Make a request
	res, err := c.send(req)
	if err != nil {
		return nil, err
	} else {
		defer res.Body.Close()
	}

	// Extract the response body
	r := new(Runner)
	if err = json.NewDecoder(res.Body).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}
