/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

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
	// Credentials / Platform overrides
}
