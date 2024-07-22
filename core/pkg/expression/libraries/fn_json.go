package libraries

import (
	"encoding/json"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func ToJSON(v ref.Val) ref.Val {
	if b, err := json.MarshalIndent(v.Value(), "", "  "); err != nil {
		return types.WrapError(err)
	} else {
		s := string(b)
		return types.String(s)
	}
}

func FromJson(v ref.Val) ref.Val {
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
