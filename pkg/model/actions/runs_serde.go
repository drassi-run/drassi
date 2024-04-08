package actions

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/model"
)

var typeRuns = reflect.TypeFor[Runs]()

func DecodeRunsHook(from reflect.Value, to reflect.Value) (any, error) {
	if !to.Type().Implements(typeRuns) {
		return from.Interface(), nil
	}
	t := to.Interface()

	m, ok := from.Interface().(map[string]any)
	if !ok || m == nil {
		return from.Interface(), nil
	}

	using, ok := m["using"].(string)
	if !ok {
		return nil, fmt.Errorf("`using` is required, and MUST be a string")
	}

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
	return from.Interface(), nil
}

func init() {
	model.RegisterDecodeHook(DecodeRunsHook)
}
