package workflows

import (
	"fmt"
	"reflect"

	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/util/reflect"
)

var typeStep = reflect.TypeFor[Step]()

func DecodeStepHook(from reflect.Value, to reflect.Value) (any, error) {
	if !from.IsValid() || !to.Type().Implements(typeStep) {
		return utilreflect.ValueOf(from), nil
	}

	f := from.Interface()
	m, ok := f.(map[string]any)
	if !ok {
		return f, nil
	}

	_, containsRun := m["run"]
	_, containsUses := m["uses"]
	if containsRun == containsUses {
		return nil, fmt.Errorf("map MUST be contains either `run` or `uses`")
	}

	t := to.Interface()
	if containsRun {
		if t == nil {
			to.Set(reflect.ValueOf(&RunStep{}))
		} else if _, ok := t.(*RunStep); !ok {
			return nil, fmt.Errorf("map contains `run` CAN'T be decode to %T", t)
		}
	}
	if containsUses {
		if t == nil {
			to.Set(reflect.ValueOf(&UsesStep{}))
		} else if _, ok := t.(*UsesStep); !ok {
			return nil, fmt.Errorf("map contains `uses` CAN'T be decode to %T", t)
		}
	}
	return m, nil
}

func init() {
	model.RegisterDecodeHook(DecodeStepHook)
}
