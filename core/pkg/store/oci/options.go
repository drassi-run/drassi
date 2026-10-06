/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import "io"

/************* MountOptions *************/

type PullOption func(*pullOptions)
type pullOptions struct {
	Policy                PullPolicy
	AuthUsername          string
	AuthPassword          string
	InsecureSkipTLSVerify bool
}

type PullPolicy int

const (
	PullIfNotPresent PullPolicy = iota
	PullAlways
	PullNever
)

func WithPullPolicy(policy PullPolicy) PullOption {
	return func(o *pullOptions) {
		o.Policy = policy
	}
}

func WithAuth(username, password string) PullOption {
	return func(o *pullOptions) {
		o.AuthUsername = username
		o.AuthPassword = password
	}
}

func WithInsecureSkipTLSVerify(insecure bool) PullOption {
	return func(o *pullOptions) {
		o.InsecureSkipTLSVerify = insecure
	}
}

/************* MountOptions *************/

type MountOption func(*mountOptions)
type mountOptions struct {
	Block    bool
	Target   string // when Block is false
	Writable bool
}

func WithBlock() MountOption {
	return func(o *mountOptions) {
		o.Block = true
	}
}

func WithTarget(dir string) MountOption {
	return func(o *mountOptions) {
		o.Target = dir
	}
}

func WithWritable(writable bool) MountOption {
	return func(o *mountOptions) {
		o.Writable = writable
	}
}

/************* ReadOptions *************/

type ReadOption func(*readOptions)
type readOptions struct {
	Subpath string
}

func WithSubpath(subpath string) ReadOption {
	return func(o *readOptions) {
		o.Subpath = subpath
	}
}

/************* utils *************/

type readCloser struct {
	io.Reader
	onClose func() error
}

func (r *readCloser) Close() error {
	if r.onClose != nil {
		return r.onClose()
	}
	return nil
}
