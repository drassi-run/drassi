package libraries

import (
	"strings"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Join(array ref.Val, separator ref.Val) string {
	if array.Type() == ref.TypeList {
		if list, ok := array.(traits.Iterable); ok {
			sep := ","
			if separator != nil {
				if s, ok := separator.(traits.Stringable); ok {
					sep = s.ToString()
				}
			}

			res := new(strings.Builder)
			it := list.Iterator()
			i := 0
			for it.HasNext() {
				_, e := it.Next()
				i++

				if i > 1 {
					res.WriteString(sep)
				}
				res.WriteString(display(e))
			}

			return res.String()
		}
	}

	if str, ok := array.(traits.Stringable); ok {
		return str.ToString()
	}
	return ""
}

func display(v ref.Val) string {
	if s, ok := v.(traits.Stringable); ok {
		return s.ToString()
	}
	return v.Type().String()
}
