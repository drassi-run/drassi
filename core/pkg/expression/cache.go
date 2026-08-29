/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package expression

import (
	"sync"

	"drassi.run/core/pkg/expression/ast"
)

type cacheNode struct {
	node ast.Node
	err  error
}

type astCache struct {
	mu sync.RWMutex
	m  map[string]*cacheNode
}

func newAstCache() *astCache {
	return &astCache{
		m: make(map[string]*cacheNode),
	}
}

func (c *astCache) Get(key string) (*cacheNode, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n, ok := c.m[key]
	return n, ok
}

func (c *astCache) Set(key string, node ast.Node, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = &cacheNode{node: node, err: err}
}

func EnableCache(e Env) Env {
	if ce, ok := e.(*cachedEnv); ok {
		return ce
	}
	return &cachedEnv{
		Env:       e,
		exprCache: newAstCache(),
		tmplCache: newAstCache(),
	}
}

type cachedEnv struct {
	Env

	exprCache *astCache
	tmplCache *astCache
}

func (e *cachedEnv) New(opts ...Option) (Env, error) {
	if d, err := e.Env.New(opts...); err != nil {
		return nil, err
	} else {
		ce := &cachedEnv{
			Env:       d,
			exprCache: e.exprCache,
			tmplCache: e.tmplCache,
		}
		return ce, nil
	}
}

func (e *cachedEnv) Parse(source string, pureExpr bool) (node ast.Node, err error) {
	cache := e.tmplCache
	if pureExpr {
		cache = e.exprCache
	}

	// check if expr already in cache
	if n, ok := cache.Get(source); ok {
		return n.node, n.err
	}

	node, err = e.Env.Parse(source, pureExpr)

	// store expr result in cache
	cache.Set(source, node, err)
	return
}
