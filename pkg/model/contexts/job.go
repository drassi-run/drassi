package contexts

// The job context contains information about the currently running job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/JobContext.cs
type Job struct {
	Container Container            `json:"container" yaml:"container"`
	Services  map[string]Container `json:"services" yaml:"services"`
	Status    ActionResult         `json:"status" yaml:"status"`
}

type Container struct {
	Id      string            `json:"id" yaml:"id"`
	Network string            `json:"network" yaml:"network"`
	Ports   map[string]string `json:"ports" yaml:"ports"`
}

// The `jobs` context is only available in reusable workflows, and can only be used to set outputs for a reusable workflow.
// https://docs.github.com/en/actions/learn-github-actions/contexts#jobs-context
type JobReusableWorkflow struct {
	Result  ActionResult      `json:"result" yaml:"result"`
	Outputs map[string]string `json:"outputs" yaml:"outputs"`
}

// The `steps` context contains information about the steps in the current job that have an `id` specified and have already run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/StepsContext.cs
type Step struct {
	Outputs    map[string]string `json:"outputs" yaml:"outputs" mapstructure:"outputs"`
	Conclusion ActionResult      `json:"conclusion" yaml:"conclusion" mapstructure:"conclusion"`
	Outcome    ActionResult      `json:"outcome" yaml:"outcome" mapstructure:"outcome"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionResult.cs
type ActionResult string

const (
	ActionResultSuccess   ActionResult = "success"
	ActionResultFailure   ActionResult = "failure"
	ActionResultCancelled ActionResult = "cancelled"
	ActionResultSkipped   ActionResult = "skipped"
)
