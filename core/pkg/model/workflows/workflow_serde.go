package workflows

import "drassi.run/core/pkg/model"

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

func (c *Concurrency) DecodeMapstructure(input any) (any, error) {
	if s, ok := model.Stringify(input); ok {
		m := map[string]any{"group": s}
		return m, nil
	}
	// process Concurrency normal way
	return input, nil
}
