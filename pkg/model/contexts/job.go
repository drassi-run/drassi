package contexts

// The job context contains information about the currently running job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
type Job struct {
	Container Container            `json:"container" yaml:"container"`
	Services  map[string]Container `json:"services" yaml:"services"`
	Status    JobStatus            `json:"status" yaml:"status"`
}

type Container struct {
	Id      string            `json:"id" yaml:"id"`
	Network string            `json:"network" yaml:"network"`
	Ports   map[string]string `json:"ports" yaml:"ports"`
}

type JobStatus string

const (
	JobStatusSuccess   = "success"
	JobStatusFailure   = "failure"
	JobStatusCancelled = "cancelled"
)

// The `jobs` context is only available in reusable workflows, and can only be used to set outputs for a reusable workflow.
// https://docs.github.com/en/actions/learn-github-actions/contexts#jobs-context
type JobReusableWorkflow struct {
	Result  JobResult         `json:"result" yaml:"result"`
	Outputs map[string]string `json:"outputs" yaml:"outputs"`
}

type JobResult string

const (
	JobResultSuccess   = "success"
	JobResultFailure   = "failure"
	JobResultCancelled = "cancelled"
	JobResultSkipped   = "skipped"
)

// The `steps` context contains information about the steps in the current job that have an `id` specified and have already run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
type Step struct {
	Outputs    map[string]string `json:"outputs" yaml:"outputs"`
	Conclusion StepResult        `json:"conclusion" yaml:"conclusion"`
	Outcome    StepResult        `json:"outcome" yaml:"outcome"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionResult.cs
type StepResult string

const (
	StepResultSuccess   StepResult = "success"
	StepResultFailure   StepResult = "failure"
	StepResultCancelled StepResult = "cancelled"
	StepResultSkipped   StepResult = "skipped"
)
