package libraries

import (
	"strings"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Join(args []ref.LazyVal) ref.Val {
	if l := len(args); l == 0 || l > 2 {
		return types.NewError("no. of args mismatch, expect [1..2], got %d", l)
	}
	array := args[0]()
	if array.Type() == ref.TypeInvalid {
		return array
	}

	if array.Type() == ref.TypeList {
		list, ok := array.(traits.Iterable)
		if !ok {
			goto next
		}

		sep := ","
		if len(args) == 2 {
			separator := args[1]()
			if separator.Type() == ref.TypeInvalid {
				return separator
			}
			if s, ok := separator.(traits.Stringable); ok {
				sep = s.ToString()
			}
		}

		return join(list, sep)
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
