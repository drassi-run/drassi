package workflows

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
			if secret, err := newEvaluable(v, toString); err != nil {
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
		if name, err := newEvaluable(s, toString); err != nil {
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
		if label, err := newEvaluable(l, toString); err != nil {
			return nil, err
		} else {
			r.Labels = append(r.Labels, label)
		}
	}
	return input, nil
}
