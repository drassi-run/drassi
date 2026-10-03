/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"io"
)

type MountOption func(*mountOptions)
type mountOptions struct {
	Writable bool
}

func WithWritable(writable bool) MountOption {
	return func(o *mountOptions) {
		o.Writable = writable
	}
}

type PullOption func(*pullOptions)
type pullOptions struct {
}

func WithAuth(username, password string) PullOption {
	return func(o *pullOptions) {

	}
}

func WithReportWriter(w io.Writer) PullOption {
	return func(o *pullOptions) {

	}
}

func WithInsecureSkipTLSVerify(insecure bool) PullOption {
	return func(o *pullOptions) {
	}
}

func WithPlatform(platform string) PullOption {
	return func(o *pullOptions) {
	}
}

type ReadOption func(*readOptions)
type readOptions struct {
	Subpath string
}

func WithSubpath(subpath string) ReadOption {
	return func(o *readOptions) {
		o.Subpath = subpath
	}
}
