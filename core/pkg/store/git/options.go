/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gitstore

import "time"

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
