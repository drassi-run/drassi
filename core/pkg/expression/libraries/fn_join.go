package libraries

import (
	"strings"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Join(array ref.Val, separator ref.LazyVal) ref.Val {
	if array.Type() == ref.TypeList {
		list, ok := array.(traits.Iterable)
		if !ok {
			goto next
		}

		dim := ","
		if separator != nil {
			delimiter := separator()
			if delimiter.Type() == ref.TypeInvalid {
				return delimiter
			}
			if s, ok := delimiter.(traits.Stringable); ok {
				dim = s.ToString()
			}
		}

		return join(list, dim)
	}

next:
	if str, ok := array.(traits.Stringable); ok {
		s := str.ToString()
		return types.String(s)
	}
	return types.String("")
}

func join(list traits.Iterable, sep string) ref.Val {
	builder := new(strings.Builder)
	it := list.Iterator()
	i := 0
	for it.HasNext() {
		_, e := it.Next()
		if e.Type() == ref.TypeInvalid {
			return e
		}

		i++
		if i > 1 {
			builder.WriteString(sep)
		}
		builder.WriteString(stringify(e))
	}

	s := builder.String()
	return types.String(s)
}

func stringify(v ref.Val) string {
	if s, ok := v.(traits.Stringable); ok {
		return s.ToString()
	}
	return v.Type().String()
}
