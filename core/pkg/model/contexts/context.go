package contexts

import (
	"context"
)

type contextKey struct{}

// Contexts are a way to access information about workflow runs, variables, runner environments, jobs, and steps.
// Each context is an object that contains properties, which can be strings or other objects.
// https://docs.github.com/en/actions/learn-github-actions/contexts#about-contexts
type Context struct {
	// The `github` context contains information about the workflow run and the event that triggered the run.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#github-contex
	Github *Github `json:"github" yaml:"github" mapstructure:"github"`

	// The `env` context contains variables that have been set in a workflow, job, or step.
	// It does not contain variables inherited by the runner process.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#env-context
	Env map[string]string `json:"env" yaml:"env" mapstructure:"env"`

	// The `vars` context contains custom configuration variables set at the organization, repository, and environment levels.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#vars-context
	Variables map[string]string `json:"vars" yaml:"vars" mapstructure:"vars"`

	// The job context contains information about the currently running job.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
	Job *Job `json:"job" yaml:"job" mapstructure:"job"`

	// The `jobs` context is only available in reusable workflows, and can only be used to set outputs for a reusable workflow.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#jobs-context
	Jobs map[string]*JobReusableWorkflow `json:"jobs" yaml:"jobs" mapstructure:"jobs"`

	// The `steps` context contains information about the steps in the current job that have an `id` specified and have already run.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
	Steps map[string]*Step `json:"steps" yaml:"steps" mapstructure:"steps"`

	// The `runner` context contains information about the runner that is executing the current job.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
	Runner *Runner `json:"runner" yaml:"runner" mapstructure:"runner"`

	// The secrets context contains the names and values of secrets that are available to a workflow run.
	// The secrets context is not available for composite actions due to security reasons.
	// If you want to pass a secret to a composite action, you need to do it explicitly as an input.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#secrets-context
	Secrets map[string]string `json:"secrets" yaml:"secrets" mapstructure:"secrets"`

	// For workflows with a matrix, the strategy context contains information about the matrix execution strategy for the current job.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#strategy-context
	Strategy *Strategy `json:"strategy" yaml:"strategy" mapstructure:"strategy"`

	// For workflows with a matrix, the matrix context contains the matrix properties defined in the workflow file that apply to the current job.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#matrix-context
	Matrix map[string]string `json:"matrix" yaml:"matrix" mapstructure:"matrix"`

	// The needs context contains outputs from all jobs that are defined as a direct dependency of the current job.
	// Note that this doesn't include implicitly dependent jobs (for example, dependent jobs of a dependent job).
	// https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
	Needs map[string]*Need `json:"needs" yaml:"needs" mapstructure:"needs"`

	// The inputs context contains input properties passed to an action, to a reusable workflow, or to a manually triggered workflow.
	// https://docs.github.com/en/actions/learn-github-actions/contexts#inputs-context
	Inputs map[string]any `json:"inputs" yaml:"inputs" mapstructure:"inputs"`
}

// For workflows with a matrix, the strategy context contains information about the matrix execution strategy for the current job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#strategy-context
type Strategy struct {
	FailFast    bool  `json:"fail-fast" yaml:"fail-fast" mapstructure:"fail-fast"`
	JobIndex    int64 `json:"job-index" yaml:"job-index" mapstructure:"job-index"`
	JobTotal    int64 `json:"job-total" yaml:"job-total" mapstructure:"job-total"`
	MaxParallel int64 `json:"max-parallel" yaml:"max-parallel" mapstructure:"max-parallel"`
}

// The needs context contains outputs from all jobs that are defined as a direct dependency of the current job.
// Note that this doesn't include implicitly dependent jobs (for example, dependent jobs of a dependent job).
// https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
type Need struct {
	Outputs map[string]string `json:"outputs" yaml:"outputs" mapstructure:"outputs"`
	Result  Result            `json:"result" yaml:"result" mapstructure:"result"`
}

type Expr struct {
	// State contains context data, used in when evaluating status check function or access context data
	// Examples for such expr:
	// name: ${{ github.actor }}
	// when: ${{ success() }}
	State *Context
}

// FromGoContext get gha context data from go context, return default value for fields if gha context data is not set
func FromGoContext(ctx context.Context) (c Context) {
	if v := ctx.Value(contextKey{}); v != nil {
		return v.(Context)
	}
	return
}

// GoContext returns a copy of parent in which the value of gha context is set.
func GoContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}
