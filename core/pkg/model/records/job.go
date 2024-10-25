package records

// The job context contains information about the currently running job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/JobContext.cs
type JobInfo struct {
	Container *Container            `json:"container" yaml:"container" actions:"container"`
	Services  map[string]*Container `json:"services" yaml:"services" actions:"services"`
	Status    Result                `json:"status" yaml:"status" actions:"status"`
}

type Container struct {
	Id      string            `json:"id" yaml:"id" actions:"id"`
	Network string            `json:"network" yaml:"network" actions:"network"`
	Ports   map[string]string `json:"ports" yaml:"ports" actions:"ports"`
}

// The `jobs` context is only available in reusable workflows, and can only be used to set outputs for a reusable workflow.
// https://docs.github.com/en/actions/learn-github-actions/contexts#jobs-context
type Job struct {
	Result  Result            `json:"result" yaml:"result" actions:"result"`
	Outputs map[string]string `json:"outputs" yaml:"outputs" actions:"outputs"`
}

func (j *Job) Status() Result {
	return j.Result
}

// The `steps` context contains information about the steps in the current job that have an `id` specified and have already run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/StepsContext.cs
type Step struct {
	Outputs    map[string]string `json:"outputs" yaml:"outputs" actions:"outputs"`
	Conclusion Result            `json:"conclusion" yaml:"conclusion" actions:"conclusion"`
	Outcome    Result            `json:"outcome" yaml:"outcome" actions:"outcome"`
}

func (s *Step) Status() Result {
	return s.Outcome
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionResult.cs
type Result string

const (
	ResultSuccess   Result = "success"
	ResultFailure   Result = "failure"
	ResultCancelled Result = "cancelled"
	ResultSkipped   Result = "skipped"
)
