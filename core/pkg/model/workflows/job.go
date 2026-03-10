/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

type Job interface {
	Base() *BaseJob
}

// ensure Job implementations
var (
	_ Job = (*NormalJob)(nil)
	_ Job = (*ReusableWorkflowCallJob)(nil)
)

// A workflow run is made up of one or more jobs. Jobs run in parallel by default. To run jobs sequentially,
// you can define dependencies on other jobs using the jobs.<job_id>.needs keyword.
// Each job runs in a fresh instance of the virtual environment specified by runs-on.
// You can run an unlimited number of jobs as long as you are within the workflow usage limits.
// For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#usage-limits.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobs
type BaseJob struct {
	// The name of the job displayed on GitHub
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idname
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Name Evaluable[string] `json:"name,omitempty" yaml:"name,omitempty" actions:"name,omitempty"`

	// You can modify the default permissions granted to the GITHUB_TOKEN,
	// adding or removing access as required, so that you only allow the minimum required access.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idpermissions
	Permissions Permissions `json:"permissions,omitempty" yaml:"permissions,omitempty" actions:"permissions,omitempty"`

	// Identifies any jobs that must complete successfully before this job will run. It can be a string or array of strings.
	// If a job fails, all jobs that need it are skipped unless the jobs use a conditional statement that causes the job to continue.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idneeds
	Needs JobNeeds `json:"needs,omitempty" yaml:"needs,omitempty" actions:"needs,omitempty"`

	// You can use the if conditional to prevent a job from running unless a condition is met.
	// You can use any supported context and expression to create a conditional.
	// Expressions in an if conditional do not require the ${{ }} syntax.
	// For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idif
	//
	// Context available: `github`, `needs`, `vars`, `inputs`
	// Special functions: `always`, `cancelled`, `success`, `failure`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	If Conditional `json:"if,omitempty" yaml:"if,omitempty" actions:"if,omitempty"`

	// A strategy creates a build matrix for your jobs. You can define different variations of an environment to run each job in.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategy
	Strategy Strategy `json:"strategy,omitempty" yaml:"strategy,omitempty" actions:"strategy,omitempty"`

	// Concurrency ensures that only a single job or workflow using the same concurrency group will run at a time.
	// A concurrency group can be any string or expression. The expression can use any context except for the secrets context.
	// You can also specify concurrency at the workflow level. When a concurrent job or workflow is queued,
	// if another job or workflow using the same concurrency group in the repository is in progress,
	// the queued job or workflow will be pending. Any previously pending job or workflow in the concurrency group will be canceled.
	// To also cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idconcurrency
	Concurrency Concurrency `json:"concurrency,omitempty" yaml:"concurrency,omitempty" actions:"concurrency,omitempty"`

	// Prevents a workflow run from failing when a job fails. Set to true to allow a workflow run to pass when this job fails.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontinue-on-error
	//
	// Context available: `github`, `needs`, `strategy`, `vars`, `matrix`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	ContinueOnError Evaluable[bool] `json:"continue-on-error,omitempty" yaml:"continue-on-error,omitempty" actions:"continue-on-error,omitempty"`
}

func (j *BaseJob) Base() *BaseJob {
	return j
}

type NormalJob struct {
	BaseJob `json:",inline" yaml:",inline" actions:",squash"`

	// The type of machine to run the job on. The machine can be either a GitHub-hosted runner, or a self-hosted runner.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idruns-on
	RunsOn Evaluable[RunsOn] `json:"runs-on,omitempty" yaml:"runs-on,omitempty" actions:"runs-on,omitempty" validate:"required"`

	// The environment that the job references.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenvironment
	Environment Evaluable[Environment] `json:"environment,omitempty" yaml:"environment,omitempty" actions:"environment,omitempty"`

	// A map of outputs for a job. Job outputs are available to all downstream jobs that depend on this job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idoutputs
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `secrets`, `steps`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Outputs Evaluable[map[string]string] `json:"outputs,omitempty" yaml:"outputs,omitempty" actions:"outputs,omitempty"`

	// A map of environment variables that are available to all steps in the job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenv
	//
	// To set custom environment variables, you need to specify the variables in the workflow file.
	// You can define environment variables for a step, job, or entire workflow using the jobs.<job_id>.steps[*].env, jobs.<job_id>.env, and env keywords.
	// For more information, see https://docs.github.com/en/actions/learn-github-actions/variables
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `vars`, `secrets`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Env Evaluable[map[string]string] `json:"env,omitempty" yaml:"env,omitempty" actions:"env,omitempty"`

	// A map of default settings that will apply to all steps in the job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_iddefaults
	Defaults Evaluable[Defaults] `json:"defaults,omitempty" yaml:"defaults,omitempty" actions:"defaults,omitempty"`

	// A job contains a sequence of tasks called steps. Steps can run commands, run setup tasks, or run an action in your repository,
	// a public repository, or an action published in a Docker registry. Not all steps run actions, but all actions run as a step.
	// Each step runs in its own process in the virtual environment and has access to the workspace and filesystem.
	// Because steps run in their own process, changes to environment variables are not preserved between steps.
	// GitHub provides built-in steps to set up and complete a job. Must contain either `uses` or `run`
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsteps
	Steps []Step `json:"steps,omitempty" yaml:"steps,omitempty" actions:"steps,omitempty"`

	// The maximum number of minutes to let a workflow run before GitHub automatically cancels it. Default: 360
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	TimeoutInMinutes Evaluable[int64] `json:"timeout-minutes,omitempty" yaml:"timeout-minutes,omitempty" actions:"timeout-minutes,omitempty"`

	// A container to run any steps in a job that don't already specify a container.
	// If you have steps that use both script and container actions,
	// the container actions will run as sibling containers on the same network with the same volume mounts.
	// If you do not set a container, all steps will run directly on the host specified by runs-on unless a step
	// refers to an action configured to run in a container.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainer
	Container Evaluable[*Container] `json:"container,omitempty" yaml:"container,omitempty" actions:"container,omitempty"`

	// Additional containers to host services for a job in a workflow. These are useful for creating databases or cache services like redis.
	// The runner on the virtual machine will automatically create a network and manage the life cycle of the service containers.
	// When you use a service container for a job or your step uses container actions, you don't need to set port information to access the service.
	// Docker automatically exposes all ports between containers on the same network.
	// When both the job and the action run in a container, you can directly reference the container by its hostname.
	// The hostname is automatically mapped to the service name.
	// When a step does not use a container action, you must access the service using localhost and bind the ports.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idservices
	Services Evaluable[map[string]*Container] `json:"services,omitempty" yaml:"services,omitempty" actions:"services,omitempty"`
}

// Each job must have an id to associate with the job.
// The key job_id is a string and its value is a map of the job's configuration data.
// You must replace <job_id> with a string that is unique to the jobs object.
// The <job_id> must start with a letter or _ and contain only alphanumeric characters, -, or _.", type: "object
// https://docs.github.com/en/actions/using-workflows/reusing-workflows#calling-a-reusable-workflow
type ReusableWorkflowCallJob struct {
	BaseJob `json:",inline" yaml:",inline" actions:",squash"`

	// The location and version of a reusable workflow file to run as a job, of the form './{path/to}/{localfile}.yml'
	// or '{owner}/{repo}/{path}/{filename}@{ref}'. {ref} can be a SHA, a release tag, or a branch name.
	// Using the commit SHA is the safest for stability and security.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_iduses
	Uses string `json:"uses,omitempty" yaml:"uses,omitempty" actions:"uses,omitempty" validate:"required"`

	// A map of inputs that are passed to the called workflow.
	// Any inputs that you pass must match the input specifications defined in the called workflow.
	// Unlike 'jobs.<job_id>.steps[*].with', the inputs you pass with 'jobs.<job_id>.with' are not be available
	// as environment variables in the called workflow. Instead, you can reference the inputs by using the inputs context.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idwith
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `inputs`, `vars`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	With Evaluable[map[string]string] `json:"with,omitempty" yaml:"with,omitempty" actions:"with,omitempty"`

	// When a job is used to call a reusable workflow, you can use 'secrets' to provide a map of secrets that are passed to the called workflow.
	// Any secrets that you pass must match the names defined in the called workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsecrets
	Secrets Evaluable[JobSecrets] `json:"secrets,omitempty" yaml:"secrets,omitempty" actions:"secrets,omitempty"`
}

// array support decode from string shorthand
type array []string

type JobNeeds = array

type JobSecrets struct {
	// A pair consisting of a string identifier for the secret and the value of the secret.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsecretssecret_id
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `secrets`, `inputs`, `vars`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Secrets map[string]string

	// Use the inherit keyword to pass all the calling workflow's secrets to the called workflow.
	// This includes all secrets the calling workflow has access to, namely organization, repository, and environment secrets.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsecretsinherit
	Inherit bool `json:"inherit,omitempty" yaml:"inherit,omitempty" actions:"inherit,omitempty"`
}

// The environment that the job references
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenvironment
type Environment struct {
	// The name of the environment configured in the repo.
	// Context available: `github`, `needs`, `strategy`, `matrix`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Name string `json:"name,omitempty" yaml:"name,omitempty" actions:"name,omitempty"`

	// A deployment URL
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `steps`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Url string `json:"url,omitempty" yaml:"url,omitempty" actions:"url,omitempty"`
}

// The type of machine to run the job on. The machine can be either a GitHub-hosted runner, or a self-hosted runner.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idruns-on
type RunsOn struct {
	// You can use runs-on to target runner groups, so that the job will execute on any runner that is a member of that group.
	// For more granular control, you can also combine runner groups with labels.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#choosing-runners-in-a-group
	Group string `json:"group,omitempty" yaml:"group,omitempty" actions:"group,omitempty"`

	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idruns-on
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Labels array `json:"labels,omitempty" yaml:"labels,omitempty" actions:"labels,omitempty"`
}

// A strategy creates a build matrix for your jobs. You can define different variations of an environment to run each job in.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategy
//
// Context available: `github`, `needs`, `vars`, `inputs`
// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
type Strategy struct {
	Matrix Evaluable[Matrix] `json:"matrix,omitempty" yaml:"matrix,omitempty" actions:"matrix,omitempty"`
	// When set to true, GitHub cancels all in-progress jobs if any matrix job fails. Default: true
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategyfail-fast
	FailFast Evaluable[bool] `json:"fail-fast,omitempty" yaml:"fail-fast,omitempty" actions:"fail-fast,omitempty"` // default: true

	// The maximum number of jobs that can run simultaneously when using a matrix job strategy.
	// By default, GitHub will maximize the number of jobs run in parallel depending on the available runners on GitHub-hosted virtual machines.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymax-parallel
	MaxParallel Evaluable[int64] `json:"max-parallel,omitempty" yaml:"max-parallel,omitempty" actions:"max-parallel,omitempty"`
}

type Matrix struct {
}
