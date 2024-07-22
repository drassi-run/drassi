package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	"math"
)

const wildcard = types.String("*")

func Index(args []ref.LazyVal) ref.Val {
	if len(args) == 0 {
		return types.NewError("missing args")
	}
	value := args[0]()
	if value.Type() == ref.TypeInvalid {
		return value
	}

	for _, arg := range args[1:] {
		idx := arg()
		if value.Type() == ref.TypeInvalid {
			return value
		}

		if wildcard.Equal(idx) {
			value = wildcardMember(value)
		} else {
			value = indexMember(value, idx)
		}
	}

	return value
}

func indexMember(value ref.Val, index ref.Val) ref.Val {
	if value.Type() == ref.TypeNull {
		return value
	}
	if indexer, ok := value.(traits.Indexer); ok {
		idxType := indexer.IndexType()
		idx := coerceIndex(index, idxType)
		if idx == nil {
			return types.NULL
		}

		return indexer.Get(idx)
	}
	return types.NULL
}

func wildcardMember(value ref.Val) ref.Val {
	if iterable, ok := value.(traits.Iterable); ok {
		it := iterable.Iterator()
		for it.HasNext() {
			_, _ = it.Next()
			// TODO implement it
		}
	}

	return types.NULL
}

func coerceIndex(index ref.Val, typ ref.Type) any {
	switch typ {
	case ref.TypeInteger:
		if n, ok := index.(traits.Numerical); ok {
			num := n.ToNumber()
			if math.IsNaN(num) {
				return 0
			}
			return int(num)
		}
	case ref.TypeString:
		if s, ok := index.(traits.Stringable); ok {
			str := s.ToString()
			return str
		}
	}
	return nil
}
