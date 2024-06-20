package actions

import (
	"fmt"
	"reflect"
	"strings"

	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/util/reflect"
)

var typeRuns = reflect.TypeFor[Runs]()

func DecodeRunsHook(from reflect.Value, to reflect.Value) (any, error) {
	if !from.IsValid() || !to.Type().Implements(typeRuns) {
		return utilreflect.ValueOf(from), nil
	}

	f := from.Interface()
	m, ok := f.(map[string]any)
	if !ok {
		return f, nil
	}

	using, ok := m["using"].(string)
	if !ok {
		return nil, fmt.Errorf("`using` is required, and MUST be a string")
	}

	t := to.Interface()
	if using == "composite" {
		if t == nil {
			to.Set(reflect.ValueOf(&CompositeRuns{}))
		} else if _, ok := t.(*CompositeRuns); !ok {
			return nil, fmt.Errorf(`map with using="%s" CAN'T be decode to %T`, using, t)
		}
	} else if using == "docker" {
		if t == nil {
			to.Set(reflect.ValueOf(&DockerRuns{}))
		} else if _, ok := t.(*DockerRuns); !ok {
			return nil, fmt.Errorf(`map with using="%s" CAN'T be decode to %T`, using, t)
		}
	} else if strings.HasPrefix(using, "node") {
		if t == nil {
			to.Set(reflect.ValueOf(&JavaScriptRuns{}))
		} else if _, ok := t.(*JavaScriptRuns); !ok {
			return nil, fmt.Errorf(`map with using="%s" CAN'T be decode to %T`, using, t)
		}
	}
	return m, nil
}

func init() {
	model.RegisterDecodeHook(DecodeRunsHook)
}
