/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	hc           *http.Client
	baseUrl      *url.URL
	headers      http.Header
	errorHandler func(int, http.Header, io.Reader) error
}

func NewClient(baseUrl string) (*Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	c := &Client{
		hc:      http.DefaultClient,
		baseUrl: u,
		headers: make(http.Header),
	}
	return c, nil
}

func (c *Client) HttpClient() *http.Client {
	return c.hc
}

func (c *Client) WithHttpClient(hc *http.Client) *Client {
	client := c.clone()
	client.hc = hc
	return client
}

func (c *Client) WithBaseUrl(baseUrl string) (*Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	client := c.clone()
	client.baseUrl = u
	return client, nil
}

func (c *Client) WithDefaultHeader(k, v string) *Client {
	client := c.clone()
	client.headers.Set(k, v)
	return client
}

func (c *Client) WithDefaultErrorHandler(h func(int, http.Header, io.Reader) error) *Client {
	client := c.clone()
	client.errorHandler = h
	return client
}

func (c *Client) clone() *Client {
	newClient := *c
	newClient.headers = maps.Clone(c.headers)
	return &newClient
}

func (c *Client) New(method, path string) *Execution {
	u := c.baseUrl.JoinPath(path)
	return &Execution{
		hc:        c.hc,
		method:    method,
		url:       u,
		queries:   u.Query(),
		headers:   maps.Clone(c.headers),
		onFailure: c.errorHandler,
	}
}

func (c *Client) Get(path string) *Execution    { return c.New(http.MethodGet, path) }
func (c *Client) Post(path string) *Execution   { return c.New(http.MethodPost, path) }
func (c *Client) Put(path string) *Execution    { return c.New(http.MethodPut, path) }
func (c *Client) Patch(path string) *Execution  { return c.New(http.MethodPatch, path) }
func (c *Client) Delete(path string) *Execution { return c.New(http.MethodDelete, path) }

type Execution struct {
	hc     *http.Client
	url    *url.URL
	method string

	queries      url.Values
	headers      http.Header
	bodyProvider func() (io.Reader, string, error)

	onSend    []func(req *http.Request) (err error)
	onReceive []func(resp *http.Response) (skip bool, err error)
	onSuccess func(body io.Reader) error
	onFailure func(code int, header http.Header, body io.Reader) error
}

func (e *Execution) SetQuery(k, v string) *Execution {
	e.queries.Set(k, v)
	return e
}

func (e *Execution) AddQuery(k, v string) *Execution {
	e.queries.Add(k, v)
	return e
}

func (e *Execution) SetHeader(k, v string) *Execution {
	e.headers.Set(k, v)
	return e
}

func (e *Execution) AddHeader(k, v string) *Execution {
	e.headers.Add(k, v)
	return e
}

func (e *Execution) WithBody(body io.Reader) *Execution {
	e.bodyProvider = func() (io.Reader, string, error) {
		return body, "", nil
	}
	return e
}

func (e *Execution) WithBodyProvider(bp func() (io.Reader, string, error)) *Execution {
	e.bodyProvider = bp
	return e
}

func (e *Execution) BeforeRequestSend(fn func(*http.Request) error) *Execution {
	e.onSend = append(e.onSend, fn)
	return e
}

func (e *Execution) AfterResponseReceive(fn func(*http.Response) (bool, error)) *Execution {
	e.onReceive = append(e.onReceive, fn)
	return e
}

func (e *Execution) OnSuccess(fn func(io.Reader) error) *Execution {
	e.onSuccess = fn
	return e
}

func (e *Execution) OnFailure(fn func(int, http.Header, io.Reader) error) *Execution {
	e.onFailure = fn
	return e
}

func (e *Execution) Do(ctx context.Context) (err error) {
	u := e.url
	u.RawQuery = e.queries.Encode()

	var (
		body        io.Reader
		contentType string
	)
	if e.bodyProvider != nil {
		if body, contentType, err = e.bodyProvider(); err != nil {
			return
		}
	}

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, e.method, u.String(), body); err != nil {
		return
	}
	req.Header = e.headers
	if contentType != "" {
		ct := req.Header.Get("Content-Type")
		if ct == "" {
			req.Header.Set("Content-Type", contentType)
		} else if matched, err := e.matchContentType(ct, contentType); err != nil {
			return err
		} else if !matched {
			return fmt.Errorf("Content-Type=%q is not match with the body (%s)", ct, contentType)
		}
	}

	for _, fn := range e.onSend {
		if err := fn(req); err != nil {
			return err
		}
	}

	var resp *http.Response
	if resp, err = e.hc.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	for _, fn := range e.onReceive {
		if skip, err := fn(resp); err != nil {
			return err
		} else if skip {
			return nil
		}
	}

	if resp.StatusCode >= 400 {
		if e.onFailure != nil {
			return e.onFailure(resp.StatusCode, resp.Header, resp.Body)
		}
		return fmt.Errorf("http failed with status code %d", resp.StatusCode)
	}

	if e.onSuccess != nil {
		return e.onSuccess(resp.Body)
	}
	return nil
}

func (e *Execution) matchContentType(x, y string) (bool, error) {
	ct1, _, err := mime.ParseMediaType(x)
	if err != nil {
		return false, err
	}

	ct2, _, err := mime.ParseMediaType(y)
	if err != nil {
		return false, err
	}

	return strings.EqualFold(ct1, ct2), nil
}

func JsonEncode(obj any) func() (io.Reader, string, error) {
	return func() (io.Reader, string, error) {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(obj); err != nil {
			return nil, "", err
		}
		return buf, "application/json", nil
	}
}

func JsonDecode(obj any) func(io.Reader) error {
	return func(body io.Reader) error {
		return json.NewDecoder(body).Decode(obj)
	}
}
