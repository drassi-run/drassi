package libraries

import (
	"encoding/json"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func ToJSON(v ref.Val) ref.Val {
	// json returns error `unsupported value` with the following values
	switch {
	case types.NEGATIVE_INF.Equal(v):
		return types.String("-Infinity")
	case types.POSITIVE_INF.Equal(v):
		return types.String("Infinity")
	case types.IsNaN(v):
		return types.String("NaN")
	}

	if b, err := json.MarshalIndent(v.Value(), "", "  "); err != nil {
		return types.WrapError(err)
	} else {
		s := string(b)
		return types.String(s)
	}
}

func FromJson(v ref.Val) ref.Val {
	str, ok := v.(traits.Stringable)
	if !ok {
		return types.NewError("unable to convert %v to traits.Stringable", v)
	}

	s := str.ToString()
	// json can't Unmarshal bellow value
	switch s {
	case "-Infinity":
		return types.NEGATIVE_INF
	case "Infinity":
		return types.POSITIVE_INF
	case "NaN":
		return types.NAN
	}

	b := []byte(s)
	var o any

	if err := json.Unmarshal(b, &o); err != nil {
		return types.WrapError(err)
	} else {
		val := types.NativeToVal(o)
		return val
	}
}
