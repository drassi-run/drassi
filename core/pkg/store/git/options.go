/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitstore

import (
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	defaultSize = 100
	defaultTTL  = 24 * time.Hour
)

type Option func(*options)
type options struct {
	size int
	ttl  time.Duration
}

func WithSize(size int) Option {
	return func(o *options) {
		o.size = size
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

type FetchOption func(*fetchOptions)
type fetchOptions struct {
	auth transport.AuthMethod
}

func WithToken(token string) FetchOption {
	return func(o *fetchOptions) {
		o.auth = &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}
}

type ReadOption func(*readOptions)
type readOptions struct {
	subpath string
	file    string
}

func WithSubpath(subpath string) ReadOption {
	return func(o *readOptions) {
		o.subpath = subpath
	}
}

func WithFile(file string) ReadOption {
	return func(o *readOptions) {
		o.file = file
	}
}
