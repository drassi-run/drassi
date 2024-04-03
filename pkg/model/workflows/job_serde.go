package workflows

import "fmt"

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
				s.Secrets[k] = secret
			}
		}
		return nil, nil
	}
	return nil, fmt.Errorf("unsupport input %s, type %T", input, input)
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

func (p *Permissions) DecodeMapstructure(input any) (any, error) {
	switch input {
	case "read-all":
		*p = Permissions{
			Actions:            PermissionsLevelRead,
			Checks:             PermissionsLevelRead,
			Contents:           PermissionsLevelRead,
			Deployments:        PermissionsLevelRead,
			Discussions:        PermissionsLevelRead,
			IdToken:            PermissionsLevelRead,
			Issues:             PermissionsLevelRead,
			Packages:           PermissionsLevelRead,
			Pages:              PermissionsLevelRead,
			PullRequests:       PermissionsLevelRead,
			RepositoryProjects: PermissionsLevelRead,
			SecurityEvents:     PermissionsLevelRead,
			Statuses:           PermissionsLevelRead,
		}
	case "write-all":
		*p = Permissions{
			Actions:            PermissionsLevelWrite,
			Checks:             PermissionsLevelWrite,
			Contents:           PermissionsLevelWrite,
			Deployments:        PermissionsLevelWrite,
			Discussions:        PermissionsLevelWrite,
			IdToken:            PermissionsLevelWrite,
			Issues:             PermissionsLevelWrite,
			Packages:           PermissionsLevelWrite,
			Pages:              PermissionsLevelWrite,
			PullRequests:       PermissionsLevelWrite,
			RepositoryProjects: PermissionsLevelWrite,
			SecurityEvents:     PermissionsLevelWrite,
			Statuses:           PermissionsLevelWrite,
		}
	default:
		// process Permissions normal way
		return input, nil
	}
	return nil, nil
}

func (r *RunsOn) DecodeMapstructure(input any) (any, error) {
	var labels []string
	switch input.(type) {
	case string:
		s := input.(string)
		labels = []string{s}
	case []string:
		labels = input.([]string)
	default:
		// process RunsOn normal way
		return input, nil
	}
	for _, l := range labels {
		if label, err := newEvaluable(l, toString); err != nil {
			return nil, err
		} else {
			r.Labels = append(r.Labels, label)
		}
	}
	return nil, nil
}

func (c *Concurrency) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		if group, err := newEvaluable(s, toString); err != nil {
			return nil, err
		} else {
			c.Group = group
			return nil, nil
		}
	}
	// process Concurrency normal way
	return input, nil
}
