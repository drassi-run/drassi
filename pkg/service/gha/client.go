package gha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	utilhttp "github.com/dungdm93/drasi/pkg/util/http"
	"golang.org/x/oauth2"
)

const (
	groupEndpoint    = "_apis/distributedtask/pools"
	runnerEndpoint   = groupEndpoint + "/%d/agents"
	sessionEndpoint  = groupEndpoint + "/%d/sessions"
	messagesEndpoint = groupEndpoint + "/%d/messages"
	apiVersion       = "6.0-preview"
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

func NewClient(ctx context.Context, serverUrl string, tokenSource oauth2.TokenSource) (*Client, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return nil, err
	}

	hc := oauth2.NewClient(ctx, tokenSource)
	c := Client{
		h:         hc,
		serverUrl: u,
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
		q.Set("api-version", apiVersion)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent.String())

	return req, nil
}

func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
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
	groups := new(ghaResponse[Group])
	if err = json.NewDecoder(res.Body).Decode(groups); err != nil {
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
	runners := new(ghaResponse[RunnerReference])
	if err = json.NewDecoder(res.Body).Decode(runners); err != nil {
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

func (c *Client) CreateSession(ctx context.Context, groupId int32, session *Session) (*Session, error) {
	// Construct request
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(session); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(sessionEndpoint, groupId)
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
	s := new(Session)
	if err = json.NewDecoder(res.Body).Decode(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Client) DeleteSession(ctx context.Context, groupId int32, sessionId string) error {
	endpoint := path.Join(fmt.Sprintf(sessionEndpoint, groupId), sessionId)
	req, err := c.NewActionsServiceRequest(ctx, http.MethodDelete, endpoint, nil, nil)
	if err != nil {
		return err
	}

	res, err := c.send(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

type GetMessageOptions struct {
	SessionId     string
	Status        RunnerStatus
	RunnerVersion string
	OS            string
	Architecture  string
	DisableUpdate bool
}

func (o *GetMessageOptions) ToQueryMap() map[string]string {
	query := make(map[string]string)
	if o.SessionId != "" {
		query["sessionId"] = o.SessionId
	}
	if o.Status != "" {
		query["status"] = string(o.Status)
	}
	if o.RunnerVersion != "" {
		query["runnerVersion"] = o.RunnerVersion
	}
	if o.OS != "" {
		query["os"] = o.OS
	}
	if o.Architecture != "" {
		query["architecture"] = o.Architecture
	}
	if o.DisableUpdate {
		query["disableUpdate"] = "true"
	}
	return query
}

func (c *Client) GetMessage(ctx context.Context, groupId int32, opts GetMessageOptions) (*Message, error) {
	// Construct request
	query := opts.ToQueryMap()
	endpoint := fmt.Sprintf(messagesEndpoint, groupId)
	req, err := c.NewActionsServiceRequest(ctx, http.MethodGet, endpoint, query, nil)
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
	message := new(Message)
	if err = json.NewDecoder(res.Body).Decode(message); err != nil {
		return nil, err
	}
	return message, nil
}
