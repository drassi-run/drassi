package workflows

import (
	"fmt"
	"reflect"

	"github.com/dungdm93/drassi/core/pkg/model"
)

var typeJob = reflect.TypeFor[Job]()

func DecodeJobHook(from reflect.Value, to reflect.Value) (any, error) {
	if !to.Type().Implements(typeJob) {
		return valueOf(from), nil
	}
	t := to.Interface()

	raw := valueOf(from)
	m, ok := raw.(map[string]any)
	if !ok || m == nil {
		return raw, nil
	}

	_, containsRunsOn := m["runs-on"]
	_, containsUses := m["uses"]

	if containsRunsOn == containsUses {
		return nil, fmt.Errorf("map MUST be contains either `runs-on` or `uses`")
	}
	if containsRunsOn {
		if t == nil {
			to.Set(reflect.ValueOf(&NormalJob{}))
		} else if _, ok := t.(*NormalJob); !ok {
			return nil, fmt.Errorf("map contains `runs-on` CAN'T be decode to %T", t)
		}
	}
	if containsUses {
		if t == nil {
			to.Set(reflect.ValueOf(&ReusableWorkflowCallJob{}))
		} else if _, ok := t.(*ReusableWorkflowCallJob); !ok {
			return nil, fmt.Errorf("map contains `uses` CAN'T be decode to %T", t)
		}
	}
	return m, nil
}

func init() {
	model.RegisterDecodeHook(DecodeJobHook)
}

func (n *JobNeeds) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		return []string{s}, nil
	} else {
		return input, nil
	}
}

func (s *JobSecrets) DecodeMapstructure(input any) (any, error) {
	if input == "inherit" {
		s.Inherit = true
		return nil, nil
	}
	if m, ok := input.(map[string]string); ok {
		s.Secrets = m
		return nil, nil
	}
	if m, ok := input.(map[string]any); ok {
		if secrets, err := castMap[string, string](m); err != nil {
			return nil, err
		} else {
			s.Secrets = secrets
			return nil, nil
		}
	}
	// process JobSecrets normal way
	return input, nil
}

func (e *Environment) DecodeMapstructure(input any) (any, error) {
	if name, ok := input.(string); ok {
		e.Name = name
		return nil, nil
	}
	// process Environment normal way
	return input, nil
}

func (r *RunsOn) setLabels(input any, rec bool) (any, error) {
	switch inp := input.(type) {
	case string:
		r.Labels = []string{inp}
		return nil, nil
	case []string:
		r.Labels = inp
		return nil, nil
	case []any:
		if labels, err := castArray[string](inp); err != nil {
			return nil, err
		} else {
			r.Labels = labels
			return nil, nil
		}
	case map[string]any:
		if rec {
			if labels, ok := inp["labels"]; ok {
				remain, err := r.setLabels(labels, false)
				if err != nil {
					return nil, err
				}
				if remain == nil { // Labels is set
					delete(inp, "labels")
				} else {
					inp["labels"] = remain
				}
			}
		}
		return input, nil
	default:
		return input, nil
	}
}

func (r *RunsOn) DecodeMapstructure(input any) (any, error) {
	return r.setLabels(input, true)
}
