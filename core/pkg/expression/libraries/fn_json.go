/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"encoding/json"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func ToJSON(v ref.Val) ref.Val {
	//TODO: go's encoding/json not support +/-Infinity and NaN, but C# does

	if b, err := json.MarshalIndent(v.Value(), "", "  "); err != nil {
		return types.WrapError(err)
	} else {
		s := string(b)
		return types.String(s)
	}
}

func FromJson(v ref.Val) ref.Val {
	//TODO: go's encoding/json not support +/-Infinity and NaN, but C# does

	s, ok := v.(traits.Stringable)
	if !ok {
		return types.NewError("unable to convert %v to traits.Stringable", v)
	}

	b := []byte(s.ToString())

	var o any
	if err := json.Unmarshal(b, &o); err != nil {
		return types.WrapError(err)
	} else {
		val := types.NativeToVal(o)
		return val
	}
}
