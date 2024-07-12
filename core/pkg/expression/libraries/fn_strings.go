package libraries

import (
	"strings"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func StartsWith(str, prefix ref.Val) bool {
	s, ok := str.(traits.Stringable)
	if !ok {
		return false
	}
	f, ok := prefix.(traits.Stringable)
	if !ok {
		return false
	}

	ss := s.ToString()
	sf := f.ToString()
	if len(ss) < len(sf) {
		return false
	}
	// GitHub ignores case when comparing strings.
	return strings.EqualFold(ss[:len(sf)], sf)
}

func EndsWith(str, suffix ref.Val) bool {
	s, ok := str.(traits.Stringable)
	if !ok {
		return false
	}
	f, ok := suffix.(traits.Stringable)
	if !ok {
		return false
	}

	// GitHub ignores case when comparing strings.
	ss := s.ToString()
	sf := f.ToString()
	if len(ss) < len(sf) {
		return false
	}
	// GitHub ignores case when comparing strings.
	return strings.EqualFold(ss[len(ss)-len(sf):], sf)
}
