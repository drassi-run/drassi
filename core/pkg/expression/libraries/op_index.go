package libraries

import (
	"math"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

const wildcard = types.String("*")

func Index(object ref.LazyVal, indexes ...ref.LazyVal) ref.Val {
	value := object()

	filterMode := false
	for _, index := range indexes {
		// short-circuit: select member from NULL always return NULL
		if ref.IsError(value) || ref.IsNull(value) {
			return value
		}

		idx := index()
		if ref.IsError(idx) {
			return idx
		}

		if !filterMode {
			if !wildcard.Equal(idx) {
				value = selectMember(value, idx)
			} else {
				value = wildcardMember(value)
				filterMode = true
			}
			continue
		}

		// In filterMode, value always is a List
		iterable, ok := value.(traits.Iterable)
		if !ok {
			return types.NULL // should never be reached here
		}

		children := make([]any, 0)
		for _, item := range iterable.Iterator() {
			if ref.IsError(item) {
				return item
			}

			err := extractMembers(item, idx, func(v any) {
				if v != nil {
					children = append(children, v)
				}
			})
			if err != nil {
				return err
			}
		}

		value = types.NewListGeneric(children)
		if len(children) == 0 {
			// short-circuit: filter members from empty List always return empty List
			return value
		}
	}

	return value
}

func selectMember(value ref.Val, index ref.Val) ref.Val {
	if indexer, ok := value.(traits.Indexer); ok {
		idxType := indexer.IndexType()
		idx := coerceIndex(index, idxType)

		if idx != nil {
			return indexer.Get(idx)
		}
	}
	return types.NULL
}

func wildcardMember(value ref.Val) ref.Val {
	if value.Type() == ref.TypeList {
		return value
	}

	if iterable, ok := value.(traits.Iterable); ok {
		list := make([]any, 0)

		for _, v := range iterable.Iterator() {
			list = append(list, v)
		}

		return types.NewListGeneric(list)
	}

	return types.NULL
}

// TODO use iter.Seq when go1.23 released
func extractMembers(value, idx ref.Val, fn func(any)) ref.Val {
	if !wildcard.Equal(idx) {
		member := selectMember(value, idx)
		if ref.IsError(member) {
			return member
		}
		fn(member.Value())
		return nil
	}

	// wildcard index on Iterable value
	if iterable, ok := value.(traits.Iterable); ok {
		for _, member := range iterable.Iterator() {
			if ref.IsError(member) {
				return member
			}
			fn(member.Value())
		}
		return nil
	}

	return nil
}

//goland:noinspection GoSwitchMissingCasesForIotaConsts
func coerceIndex(index ref.Val, typ ref.Type) any {
	switch typ {
	case ref.TypeInteger:
		if n, ok := index.(traits.Numerical); ok {
			num := n.ToNumber()
			if math.IsNaN(num) || num < 0 || num > math.MaxInt {
				return nil
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
