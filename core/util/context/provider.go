/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xcontext

import "context"

type Provider interface {
	Context() context.Context
}

func NewStaticProvider(ctx context.Context) Provider {
	return &staticProvider{ctx: ctx}
}

type staticProvider struct {
	ctx context.Context
}

func (p *staticProvider) Context() context.Context {
	return p.ctx
}
