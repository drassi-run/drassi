/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
			if ref.IsError(delimiter) {
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

	i := 0
	for _, e := range list.Items() {
		if ref.IsError(e) {
			return e
		}

		if i++; i > 1 {
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
