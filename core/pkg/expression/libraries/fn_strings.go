package libraries

import (
	"strings"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func StartsWith(str, prefix ref.Val) ref.Val {
	s, ok := str.(traits.Stringable)
	if !ok {
		return types.FALSE
	}
	f, ok := prefix.(traits.Stringable)
	if !ok {
		return types.FALSE
	}

	ss := s.ToString()
	sf := f.ToString()
	if len(ss) < len(sf) {
		return types.FALSE
	}
	// GitHub ignores case when comparing strings.
	r := strings.EqualFold(ss[:len(sf)], sf)
	return types.Boolean(r)
}

func EndsWith(str, suffix ref.Val) ref.Val {
	s, ok := str.(traits.Stringable)
	if !ok {
		return types.FALSE
	}
	f, ok := suffix.(traits.Stringable)
	if !ok {
		return types.FALSE
	}

	// GitHub ignores case when comparing strings.
	ss := s.ToString()
	sf := f.ToString()
	if len(ss) < len(sf) {
		return types.FALSE
	}
	// GitHub ignores case when comparing strings.
	r := strings.EqualFold(ss[len(ss)-len(sf):], sf)
	return types.Boolean(r)
}
