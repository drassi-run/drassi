package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"drassi.run/core/pkg/expression/types/ref"
)

type String string

func (s String) Type() ref.Type {
	return ref.TypeString
}

func (s String) Value() any {
	return string(s)
}

func (s String) Equal(other ref.Val) bool {
	o, ok := other.(String)
	// GitHub ignores case when comparing strings.
	return ok && strings.EqualFold(string(s), string(o))
}

func (s String) ToBoolean() bool {
	return len(s) > 0
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/Sdk/ExpressionUtility.cs#L193
func (s String) ToNumber() float64 {
	str := strings.TrimSpace(string(s))

	// empty string returns 0
	if len(str) == 0 {
		return 0
	}

	// Not accept underscore '_'
	if strings.ContainsRune(str, '_') {
		return math.NaN()
	}

	if f := s.parseInf(str); f != nil {
		return *f
	}

	// * Only accept 0x & 0o prefix, not uppercase 0X & 0O
	// * Not accept sign prefix, e.g -0x1A, +0o37
	// * Not accept binary number e.g 0b10
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0o") {
		// Parse Octal & Hex into int32 (NOT int64). Out-of-range value treat as NaN
		if i, err := strconv.ParseInt(str, 0, 32); err == nil {
			return float64(i)
		}
	} else {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}
	}

	// otherwise NaN
	return math.NaN()
}

func (s String) parseInf(str string) *float64 {
	if len(s) < 3 {
		return nil
	}

	sign := 1
	if str[0] == '-' || str[0] == '+' {
		if str[0] == '-' {
			sign = -1
		}
		str = str[1:]
	}

	if len(str) < 3 || !strings.EqualFold(str[:3], "inf") {
		// It's not possible is an inf
		return nil
	}

	// case-sensitive comparing
	if str == "Infinity" {
		f := math.Inf(sign)
		return &f
	}

	// Maybe valid inf string in go, e.g "Inf", "-Inf" but not valid in JS
	// => returns NaN
	f := math.NaN()
	return &f
}

func (s String) ToString() string {
	return string(s)
}

func (s String) Compare(other ref.Val) (int, error) {
	o, ok := other.(String)
	if !ok {
		return 0, fmt.Errorf("%s vs. %s: %w", s.Type(), other.Type(), errUncomparable)
	}

	// GitHub ignores case when comparing strings.
	// TODO: implement strings.CompareFold(s1, s2) because not always map 1:1 uppercase to lowercase
	// e.g in Greek uppercase sigma (Σ) have 2 lowercases (σ, ς)
	ls := strings.ToLower(string(s))
	lo := strings.ToLower(string(o))
	return strings.Compare(ls, lo), nil
}
