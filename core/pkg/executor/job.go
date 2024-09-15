package executor

import "drassi.run/core/pkg/model/workflows"

type JobRun struct {
	Id   string
	Uid  string
	Name workflows.Evaluable[string]

	Container workflows.Evaluable[*workflows.Container]
	Services  workflows.Evaluable[map[string]*workflows.Container]

	Env      workflows.Evaluable[map[string]string]
	Steps    []StepRun
	Outputs  workflows.Evaluable[map[string]string]
	Defaults workflows.Evaluable[workflows.Defaults]
	// Environment
}
