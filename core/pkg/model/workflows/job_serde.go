package workflows

import (
	"fmt"
	"reflect"

	"github.com/dungdm93/drassi/core/pkg/model"
)

var typeJob = reflect.TypeFor[Job]()

func DecodeJobHook(from reflect.Value, to reflect.Value) (any, error) {
	if !to.Type().Implements(typeJob) {
		return from.Interface(), nil
	}
	t := to.Interface()

	m, ok := from.Interface().(map[string]any)
	if !ok || m == nil {
		return from.Interface(), nil
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
	return from.Interface(), nil
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
		for k, v := range m {
			if secret, err := NewEvaluable(v, toString); err != nil {
				return nil, err
			} else {
				if s.Secrets == nil {
					s.Secrets = make(map[string]Evaluable[string])
				}
				s.Secrets[k] = secret
			}
		}
		return nil, nil
	}
	// process JobSecrets normal way
	return input, nil
}

func (e *Environment) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		if name, err := NewEvaluable(s, toString); err != nil {
			return nil, err
		} else {
			e.Name = name
			return nil, nil
		}
	}
	// process Environment normal way
	return input, nil
}

func (r *RunsOn) DecodeMapstructure(input any) (any, error) {
	var labels []string
	switch i := input.(type) {
	case string:
		labels = []string{i}
		input = nil
	case []string:
		labels = i
		input = nil
	case map[string]any:
		if lb, ok := i["labels"]; ok {
			switch l := lb.(type) {
			case string:
				labels = []string{l}
			case []string:
				labels = l
			default:
				return input, nil
			}
			delete(i, "labels")
		}
	}
	for _, l := range labels {
		if label, err := NewEvaluable(l, toString); err != nil {
			return nil, err
		} else {
			r.Labels = append(r.Labels, label)
		}
	}
	return input, nil
}
