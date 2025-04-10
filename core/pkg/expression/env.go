/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package expression

import (
	"maps"

	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/util/error"
)

type Env interface {
	New(opts ...Option) (Env, error)
	Parse(source string, pureExpr bool) (node ast.Node, err error)
	Bind(node ast.Node) (prog ref.LazyVal, err error)
	Execute(prog ref.LazyVal) (result any, err error)
}

func NewEnv(opts ...Option) (Env, error) {
	e := &env{
		maxError: 32,
		// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/ExpressionConstants.cs#L29
		// Larger value is used because ANTLR have some internal nodes
		maxDepth: 512,
		// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/ExpressionConstants.cs#L30
		maxLength: 21_000,

		cacheEnabled: true,

		exprCache: make(map[string]*cacheNode),
		tmplCache: make(map[string]*cacheNode),

		alias:     make(map[string]string),
		variables: make(map[string]ref.Val),
		functions: make(map[string]Function),
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	return e, nil
}

type cacheNode struct {
	node ast.Node
	err  error
}

type env struct {
	maxError     int
	maxDepth     int
	maxLength    int
	cacheEnabled bool

	exprCache map[string]*cacheNode
	tmplCache map[string]*cacheNode

	alias     map[string]string
	variables map[string]ref.Val
	functions map[string]Function
}

func (e *env) New(opts ...Option) (Env, error) {
	n := &env{
		maxError:     e.maxError,
		maxDepth:     e.maxDepth,
		maxLength:    e.maxLength,
		cacheEnabled: e.cacheEnabled,

		exprCache: e.exprCache,
		tmplCache: e.tmplCache,

		alias:     maps.Clone(e.alias),
		variables: maps.Clone(e.variables),
		functions: maps.Clone(e.functions),
	}

	for _, opt := range opts {
		if err := opt(n); err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (e *env) Parse(source string, pureExpr bool) (node ast.Node, err error) {
	defer xerror.Recover(&err)

	cache := e.tmplCache
	if pureExpr {
		cache = e.exprCache
	}

	// check if expr already in cache
	if e.cacheEnabled {
		if n, ok := cache[source]; ok {
			return n.node, n.err
		}
	}

	// Parse expr
	opt := ast.Option{
		MaxError:  e.maxError,
		MaxDepth:  e.maxDepth,
		MaxLength: e.maxLength,
	}
	node, err = ast.Parse(source, pureExpr, opt)

	// store expr result in cache
	if e.cacheEnabled {
		cache[source] = &cacheNode{node: node, err: err}
	}
	return
}

func (e *env) Bind(node ast.Node) (prog ref.LazyVal, err error) {
	defer xerror.Recover(&err)

	b := binder{env: e}
	return b.Bind(node)
}

func (e *env) Execute(prog ref.LazyVal) (result any, err error) {
	defer xerror.Recover(&err)

	val := prog()
	if err, ok := val.(error); ok {
		return nil, err
	}
	return val.Value(), nil
}
