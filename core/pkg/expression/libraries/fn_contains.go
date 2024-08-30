package libraries

import (
	"strings"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Contains(search ref.Val, item ref.Val) ref.Val {
	// search is an array, check it contains item element
	if search.Type() == ref.TypeList {
		if list, ok := search.(traits.Iterable); ok {
			return contains(list, item)
		}
	}

	s, ok := search.(traits.Stringable)
	if !ok {
		return types.FALSE
	}
	i, ok := item.(traits.Stringable)
	if !ok {
		return types.FALSE
	}

	// Both search, item are convertable to string, check if item is a substring of search.
	// GitHub ignores case when comparing strings.
	ss := strings.ToLower(s.ToString())
	si := strings.ToLower(i.ToString())
	r := strings.Contains(ss, si)
	return types.Boolean(r)
}

func contains(list traits.Iterable, item ref.Val) ref.Val {
	for _, e := range list.Items() {
		if equalWeak(e, item) {
			return types.TRUE
		}
	}
	return types.FALSE
}
