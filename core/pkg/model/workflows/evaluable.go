/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"fmt"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

const (
	OpenExpression  = "${{"
	CloseExpression = "}}"
)

// Conditional is Evaluable[bool] type used by `if`, `pre-if` and `post-if`.
// The `${{ }}` expression syntax is optional and can be omitted.
// GitHub Actions always evaluates it as an expression.
type Conditional string

type Evaluable[R any] = Token

type Unraveler interface {
	UnravelLiteral(val any) (ref.Val, error)
	UnravelExpression(expr string, pure bool) (ref.Val, error)
	UnravelSequence(seq []Token) ([]ref.Val, error)
	UnravelMapping(pairs [][2]Token) (map[string]ref.Val, error)
}

type Token interface {
	Unravel(Unraveler) (ref.Val, error)
}

type literalToken struct {
	value any
}

func (l *literalToken) Unravel(u Unraveler) (ref.Val, error) {
	return u.UnravelLiteral(l.value)
}

func NewLiteralToken(value any) Token {
	return &literalToken{value: value}
}

type expressionToken string

func (e expressionToken) Unravel(u Unraveler) (ref.Val, error) {
	return u.UnravelExpression(string(e), false)
}

func NewExpressionToken(expr string) Token {
	e := expressionToken(expr)
	return e
}

type sequenceToken []Token

func (s sequenceToken) Unravel(u Unraveler) (ref.Val, error) {
	if val, err := u.UnravelSequence(s); err != nil {
		return nil, err
	} else {
		return types.NewListGeneric(val), nil
	}
}

func NewSequenceToken(seq []Token) Token {
	e := sequenceToken(seq)
	return e
}

type mappingToken [][2]Token

func (m mappingToken) Unravel(u Unraveler) (ref.Val, error) {
	if val, err := u.UnravelMapping(m); err != nil {
		return nil, err
	} else {
		return types.NewMapGeneric(val), nil
	}
}

func NewMappingToken(pairs [][2]Token) Token {
	e := mappingToken(pairs)
	return e
}

type squashMappingToken []Token

func (t squashMappingToken) Unravel(u Unraveler) (ref.Val, error) {
	r := make(map[string]ref.Val)

	for _, token := range t {
		if res, err := token.Unravel(u); err != nil {
			return nil, err
		} else if iter, ok := res.(traits.Iterable); !ok {
			return nil, fmt.Errorf("expected token return a map, got %T", res)
		} else {
			for k, v := range iter.Items() {
				if s, ok := k.(traits.Stringable); ok {
					r[s.ToString()] = v
				} else {
					return nil, fmt.Errorf("expected map key to be stringable, got %T", k)
				}
			}
		}
	}

	return types.NewMapGeneric(r), nil
}

func NewSquashMappingToken(tokens ...Token) Token {
	if len(tokens) == 0 {
		return nil
	}
	return squashMappingToken(tokens)
}

func Expression(token Token) (string, bool) {
	if expr, ok := token.(expressionToken); ok {
		return string(expr), ok
	}
	return "", false
}
