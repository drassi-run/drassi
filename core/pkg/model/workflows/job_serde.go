package workflows

import (
	"fmt"
	"reflect"

	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/util/reflect"
)

var typeJob = reflect.TypeFor[Job]()

func DecodeJobHook(from reflect.Value, to reflect.Value) (any, error) {
	if !from.IsValid() || !to.Type().Implements(typeJob) {
		return utilreflect.ValueOf(from), nil
	}

	f := from.Interface()
	m, ok := model.ObjectStringify(f)
	if !ok {
		return f, nil
	}

	_, containsRunsOn := m["runs-on"]
	_, containsUses := m["uses"]
	if containsRunsOn == containsUses {
		return nil, fmt.Errorf("map MUST be contains either `runs-on` or `uses`")
	}

	t := to.Interface()
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
	if s, ok := model.Stringify(input); ok {
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
		if secrets, err := utilreflect.CastMap[string, string](m); err != nil {
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
	if name, ok := model.Stringify(input); ok {
		e.Name = name
		return nil, nil
	}
	// process Environment normal way
	return input, nil
}

func (r *RunsOn) setLabels(input any, rec bool) (any, error) {
	if s, ok := model.Stringify(input); ok {
		r.Labels = []string{s}
		return nil, nil
	}
	if l, ok := model.ListStringify(input); ok {
		r.Labels = l
		return nil, nil
	}
	if m, ok := model.ObjectStringify(input); ok {
		if rec {
			if labels, ok := m["labels"]; ok {
				remain, err := r.setLabels(labels, false)
				if err != nil {
					return nil, err
				}
				if remain == nil { // Labels is set
					delete(m, "labels")
				} else {
					m["labels"] = remain
				}
			}
		}
		return m, nil
	}
	return input, nil
}

func (r *RunsOn) DecodeMapstructure(input any) (any, error) {
	return r.setLabels(input, true)
}
