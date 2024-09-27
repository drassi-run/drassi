package workflows

import (
	"fmt"
	"maps"
)

const (
	OpenExpression  = "${{"
	CloseExpression = "}}"
)

// Conditional is Evaluable[bool] type used by `if`, `pre-if` and `post-if`.
// The `${{ }}` expression syntax is optional and can be omitted.
// GitHub Actions always evaluates it as an expression.
type Conditional string

type Evaluable[R any] Token

type Unraveler interface {
	UnravelLiteral(val any) (any, error)
	UnravelExpression(expr string, pure bool) (any, error)
	UnravelSequence(seq []Token) (any, error)
	UnravelMapping(pairs [][2]Token) (any, error)
}

type Token interface {
	Unravel(Unraveler) (any, error)
}

type literalToken struct {
	value any
}

func (l *literalToken) Unravel(u Unraveler) (any, error) {
	return u.UnravelLiteral(l.value)
}

func NewLiteralToken(value any) Token {
	return &literalToken{value: value}
}

type expressionToken string

func (e expressionToken) Unravel(u Unraveler) (any, error) {
	return u.UnravelExpression(string(e), false)
}

func NewExpressionToken(expr string) Token {
	e := expressionToken(expr)
	return e
}

type sequenceToken []Token

func (s sequenceToken) Unravel(u Unraveler) (any, error) {
	return u.UnravelSequence(s)
}

func NewSequenceToken(seq []Token) Token {
	e := sequenceToken(seq)
	return e
}

type mappingToken [][2]Token

func (m mappingToken) Unravel(u Unraveler) (any, error) {
	return u.UnravelMapping(m)
}

func NewMappingToken(pairs [][2]Token) Token {
	e := mappingToken(pairs)
	return e
}

type squashMappingToken []Token

func (t squashMappingToken) Unravel(u Unraveler) (any, error) {
	r := make(map[string]any)

	for _, token := range t {
		if res, err := token.Unravel(u); err != nil {
			return nil, err
		} else if m, ok := res.(map[string]any); ok {
			maps.Copy(r, m)
		} else {
			return nil, fmt.Errorf("expected token return a map[string]any, got %T", res)
		}
	}

	return r, nil
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
