/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xhttp

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

type ClientTestSuite struct {
	suite.Suite
	c          *Client
	mux        *http.ServeMux
	mockServer *httptest.Server
}

func (s *ClientTestSuite) SetupTest() {
	s.mux = http.NewServeMux()
	s.mockServer = httptest.NewServer(s.mux)

	var err error
	s.c, err = NewClient(s.mockServer.URL)
	assert.NoError(s.T(), err)
}

func (s *ClientTestSuite) TearDownTest() {
	s.mockServer.Close()
}

func (s *ClientTestSuite) TestHeader() {
	t, count := s.T(), 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		count++

		switch count {
		case 1:
			assert.Equal(t, []string{"bar"}, r.Header.Values("X-Foo"))
		case 2:
			assert.Equal(t, []string{"bar", "baz"}, r.Header.Values("X-Foo"))
		default:
			t.Failed()
		}
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/header", handler)

	r := s.c.Get("/header").SetHeader("X-Foo", "bar")

	err := r.Do(t.Context())
	assert.NoError(t, err)

	r.AddHeader("X-Foo", "baz")
	err = r.Do(t.Context())
	assert.NoError(t, err)
}

func (s *ClientTestSuite) TestQuery() {
	t, count := s.T(), 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		count++
		query := r.URL.Query()

		switch count {
		case 1:
			assert.Equal(t, []string{"bar"}, query["q"])
		case 2:
			assert.Equal(t, []string{"bar", "baz"}, query["q"])
		default:
			t.Failed()
		}
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/query", handler)

	r := s.c.Get("/query").SetQuery("q", "bar")

	err := r.Do(t.Context())
	assert.NoError(t, err)

	r.AddQuery("q", "baz")
	err = r.Do(t.Context())
	assert.NoError(t, err)
}

func (s *ClientTestSuite) TestBody() {
	t := s.T()
	handler := func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "foobar", string(b))
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/body", handler)

	e := s.c.Post("/body").
		WithBody(strings.NewReader("foobar")).
		Do(t.Context())
	assert.NoError(t, e)

	e = s.c.Post("/body").
		WithBodyProvider(func() (io.Reader, string, error) {
			return strings.NewReader("foobar"), "", nil
		}).
		Do(t.Context())
	assert.NoError(t, e)
}

func (s *ClientTestSuite) TestBodyWithContentType() {
	t := s.T()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/body", handler)

	e := s.c.Post("/body").
		SetHeader("Content-Type", "application/json").
		WithBodyProvider(func() (io.Reader, string, error) {
			return http.NoBody, "application/json", nil
		}).
		Do(t.Context())
	assert.NoError(t, e)

	e = s.c.Post("/body").
		SetHeader("Content-Type", "application/xml").
		WithBodyProvider(func() (io.Reader, string, error) {
			return http.NoBody, "application/json", nil
		}).
		Do(t.Context())
	assert.ErrorContains(t, e, `Content-Type="application/xml" is not match with the body (application/json)`)

	e = s.c.Post("/body").
		SetHeader("Content-Type", "application/xml; charset=utf-8").
		WithBodyProvider(func() (io.Reader, string, error) {
			return http.NoBody, "application/json", nil
		}).
		Do(t.Context())
	assert.ErrorContains(t, e, `Content-Type="application/xml; charset=utf-8" is not match with the body (application/json)`)

	e = s.c.Post("/body").
		SetHeader("Content-Type", "application/xml").
		WithBodyProvider(func() (io.Reader, string, error) {
			return http.NoBody, "application/json; charset=utf-8", nil
		}).
		Do(t.Context())
	assert.ErrorContains(t, e, `Content-Type="application/xml" is not match with the body (application/json; charset=utf-8)`)
}

type S struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (s *ClientTestSuite) TestJsonEncode() {
	t := s.T()
	data := &S{Name: "drassi", Age: 18}

	handler := func(w http.ResponseWriter, r *http.Request) {
		d := new(S)
		err := json.NewDecoder(r.Body).Decode(d)
		assert.NoError(t, err)
		assert.Equal(t, data, d)

		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/body", handler)

	e := s.c.Post("/body").
		WithBodyProvider(JsonEncode(data)).
		Do(t.Context())
	assert.NoError(t, e)
}

func (s *ClientTestSuite) TestJsonDecode() {
	t := s.T()
	data := &S{Name: "drassi", Age: 18}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(data)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}
	s.mux.HandleFunc("/body", handler)

	var d S
	e := s.c.Post("/body").
		OnSuccess(JsonDecode(&d)).
		Do(t.Context())
	assert.NoError(t, e)
	assert.Equal(t, data, &d)
}

func (s *ClientTestSuite) TestOnFailure() {
	t := s.T()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
	s.mux.HandleFunc("/on-failure", handler)

	e := s.c.Get("/on-failure").
		Do(t.Context())
	assert.ErrorContains(t, e, "404")

	e = s.c.Get("/on-failure").
		OnFailure(func(code int, header http.Header, reader io.Reader) error {
			if code == http.StatusNotFound {
				return nil
			}
			return fmt.Errorf("unexpected status code: %d", code)
		}).Do(t.Context())
	assert.NoError(t, e)

	e = s.c.Get("/on-failure").
		OnFailure(func(code int, header http.Header, reader io.Reader) error {
			return fmt.Errorf("unexpected status code: %d", code)
		}).Do(t.Context())
	assert.ErrorContains(t, e, "unexpected status code: 404")
}
