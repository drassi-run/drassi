package libraries

import (
	"encoding/json"
	"fmt"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func ToJSON(v ref.Val) (string, error) {
	if b, err := json.MarshalIndent(v.Value(), "", "  "); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func FromJson(v ref.Val) (ref.Val, error) {
	s, ok := v.(traits.Stringable)
	if !ok {
		return nil, fmt.Errorf("unable to convert %v to traits.Stringable", v)
	}

	b := []byte(s.ToString())
	var o any

	if err := json.Unmarshal(b, &o); err != nil {
		return nil, err
	} else {
		val := types.NativeToVal(o)
		return val, nil
	}
}
