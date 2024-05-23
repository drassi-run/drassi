package workflows

import (
	"fmt"
	"reflect"

	"github.com/dungdm93/drassi/core/pkg/model"
)

var typeStep = reflect.TypeFor[Step]()

func DecodeStepHook(from reflect.Value, to reflect.Value) (any, error) {
	if !to.Type().Implements(typeStep) {
		return valueOf(from), nil
	}
	t := to.Interface()

	raw := valueOf(from)
	m, ok := raw.(map[string]any)
	if !ok || m == nil {
		return raw, nil
	}

	_, containsRun := m["run"]
	_, containsUses := m["uses"]

	if containsRun == containsUses {
		return nil, fmt.Errorf("map MUST be contains either `run` or `uses`")
	}
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
