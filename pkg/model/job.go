package model

// A workflow run is made up of one or more jobs. Jobs run in parallel by default. To run jobs sequentially,
// you can define dependencies on other jobs using the jobs.<job_id>.needs keyword.
// Each job runs in a fresh instance of the virtual environment specified by runs-on.
// You can run an unlimited number of jobs as long as you are within the workflow usage limits.
// For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#usage-limits.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobs
type Job struct {
	// The name of the job displayed on GitHub
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idname
	Name Evaluable[string] `json:"name,omitempty" yaml:"name,omitempty"`

	// You can modify the default permissions granted to the GITHUB_TOKEN,
	// adding or removing access as required, so that you only allow the minimum required access.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idpermissions
	Permissions Permissions `json:"permissions,omitempty" yaml:"permissions,omitempty"`

	// Identifies any jobs that must complete successfully before this job will run. It can be a string or array of strings.
	// If a job fails, all jobs that need it are skipped unless the jobs use a conditional statement that causes the job to continue.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idneeds
	Needs []string `json:"needs,omitempty" yaml:"needs,omitempty"`

	// You can use the if conditional to prevent a job from running unless a condition is met.
	// You can use any supported context and expression to create a conditional.
	// Expressions in an if conditional do not require the ${{ }} syntax.
	// For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idif
	If Evaluable[bool] `json:"if,omitempty" yaml:"if,omitempty"`

	// A strategy creates a build matrix for your jobs. You can define different variations of an environment to run each job in.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategy
	Strategy Strategy `json:"strategy,omitempty" yaml:"strategy,omitempty"`

	// Concurrency ensures that only a single job or workflow using the same concurrency group will run at a time.
	// A concurrency group can be any string or expression. The expression can use any context except for the secrets context.
	// You can also specify concurrency at the workflow level. When a concurrent job or workflow is queued,
	// if another job or workflow using the same concurrency group in the repository is in progress,
	// the queued job or workflow will be pending. Any previously pending job or workflow in the concurrency group will be canceled.
	// To also cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idconcurrency
	Concurrency Concurrency `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`

	// Prevents a workflow run from failing when a job fails. Set to true to allow a workflow run to pass when this job fails.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontinue-on-error
	ContinueOnError Evaluable[bool] `json:"continue-on-error,omitempty" yaml:"continue-on-error,omitempty"`
}

type JobNormal struct {
	Job

	// The type of machine to run the job on. The machine can be either a GitHub-hosted runner, or a self-hosted runner.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idruns-on
	RunsOn RunsOn `json:"runs-on,omitempty" yaml:"runs-on,omitempty" validate:"required"`

	// The environment that the job references.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenvironment
	Environment Environment `json:"environment,omitempty" yaml:"environment,omitempty"`

	// A map of outputs for a job. Job outputs are available to all downstream jobs that depend on this job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idoutputs
	Outputs JobOutputs `json:"outputs,omitempty" yaml:"outputs,omitempty"`

	// A map of environment variables that are available to all steps in the job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenv
	Env Env `json:"env,omitempty" yaml:"env,omitempty"`

	// A map of default settings that will apply to all steps in the job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_iddefaults
	Defaults Defaults `json:"defaults,omitempty" yaml:"defaults,omitempty"`

	// A job contains a sequence of tasks called steps. Steps can run commands, run setup tasks, or run an action in your repository,
	// a public repository, or an action published in a Docker registry. Not all steps run actions, but all actions run as a step.
	// Each step runs in its own process in the virtual environment and has access to the workspace and filesystem.
	// Because steps run in their own process, changes to environment variables are not preserved between steps.
	// GitHub provides built-in steps to set up and complete a job. Must contain either `uses` or `run`
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsteps
	Steps []Step `json:"steps,omitempty" yaml:"steps,omitempty"`

	// The maximum number of minutes to let a workflow run before GitHub automatically cancels it. Default: 360
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes
	TimeoutMinutes Evaluable[int64] `json:"timeout-minutes,omitempty" yaml:"timeout-minutes,omitempty"`

	// A container to run any steps in a job that don't already specify a container.
	// If you have steps that use both script and container actions,
	// the container actions will run as sibling containers on the same network with the same volume mounts.
	// If you do not set a container, all steps will run directly on the host specified by runs-on unless a step
	// refers to an action configured to run in a container.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idcontainer
	Container Container `json:"container,omitempty" yaml:"container,omitempty"`

	// Additional containers to host services for a job in a workflow. These are useful for creating databases or cache services like redis.
	// The runner on the virtual machine will automatically create a network and manage the life cycle of the service containers.
	// When you use a service container for a job or your step uses container actions, you don't need to set port information to access the service.
	// Docker automatically exposes all ports between containers on the same network.
	// When both the job and the action run in a container, you can directly reference the container by its hostname.
	// The hostname is automatically mapped to the service name.
	// When a step does not use a container action, you must access the service using localhost and bind the ports.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idservices
	Services map[string]Container `json:"services,omitempty" yaml:"services,omitempty"`
}

// Each job must have an id to associate with the job.
// The key job_id is a string and its value is a map of the job's configuration data.
// You must replace <job_id> with a string that is unique to the jobs object.
// The <job_id> must start with a letter or _ and contain only alphanumeric characters, -, or _.", type: "object
// https://docs.github.com/en/actions/using-workflows/reusing-workflows#calling-a-reusable-workflow
type JobReusableWorkflowCall struct {
	Job

	// The location and version of a reusable workflow file to run as a job, of the form './{path/to}/{localfile}.yml'
	// or '{owner}/{repo}/{path}/{filename}@{ref}'. {ref} can be a SHA, a release tag, or a branch name.
	// Using the commit SHA is the safest for stability and security.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_iduses
	Uses string `json:"uses,omitempty" yaml:"uses,omitempty" validate:"required"`

	// A map of inputs that are passed to the called workflow.
	// Any inputs that you pass must match the input specifications defined in the called workflow.
	// Unlike 'jobs.<job_id>.steps[*].with', the inputs you pass with 'jobs.<job_id>.with' are not be available
	// as environment variables in the called workflow. Instead, you can reference the inputs by using the inputs context.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idwith
	With With `json:"with,omitempty" yaml:"with,omitempty"`

	// When a job is used to call a reusable workflow, you can use 'secrets' to provide a map of secrets that are passed to the called workflow.
	// Any secrets that you pass must match the names defined in the called workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsecrets
	Secrets JobSecrets `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

type Step struct {
	// A unique identifier for the step. You can use the id to reference the step in contexts.
	// For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsid
	Id string `json:"id,omitempty" yaml:"id,omitempty"`

	// A name for your step to display on GitHub.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsname
	Name Evaluable[string] `json:"name,omitempty" yaml:"name,omitempty"`

	// You can use the if conditional to prevent a step from running unless a condition is met.
	// You can use any supported context and expression to create a conditional.
	// Expressions in an if conditional do not require the ${{ }} syntax.
	// For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsif
	If Evaluable[bool] `json:"if,omitempty" yaml:"if,omitempty"`

	// A map of the input parameters defined by the action. Each input parameter is a key/value pair.
	// Input parameters are set as environment variables. The variable is prefixed with INPUT_ and converted to upper case.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepswith
	With With `json:"with,omitempty" yaml:"with,omitempty"`

	// Sets environment variables for steps to use in the virtual environment. You can also set environment variables for the entire workflow or a job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsenv
	Env Env `json:"env,omitempty" yaml:"env,omitempty"`

	// Prevents a job from failing when a step fails. Set to true to allow a job to pass when this step fails.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepscontinue-on-error
	ContinueOnError Evaluable[bool] `json:"continue-on-error,omitempty" yaml:"continue-on-error,omitempty"`

	// The maximum number of minutes to run the step before killing the process.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes
	TimeoutMinutes Evaluable[int64] `json:"timeout-minutes,omitempty" yaml:"timeout-minutes,omitempty"`
}

type StepUses struct {
	Step

	// Selects an action to run as part of a step in your job. An action is a reusable unit of code.
	// You can use an action defined in the same repository as the workflow, a public repository, or in a published Docker container image (https://hub.docker.com/).
	// We strongly recommend that you include the version of the action you are using by specifying a Git ref, SHA, or Docker tag number.
	// If you don't specify a version, it could break your workflows or cause unexpected behavior when the action owner publishes an update.
	// - Using the commit SHA of a released action version is the safest for stability and security.
	// - Using the specific major action version allows you to receive critical fixes and security patches while still maintaining compatibility.
	//   It also assures that your workflow should still work.
	// - Using the master branch of an action may be convenient, but if someone releases a new major version with a breaking change, your workflow could break.
	// Some actions require inputs that you must set using the with keyword. Review the action's README file to determine the inputs required.
	// Actions are either JavaScript files or Docker containers. If the action you're using is a Docker container you must run the job in a Linux virtual environment.
	// For more details, see https://help.github.com/en/articles/virtual-environments-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsuses
	Uses string `json:"uses,omitempty" yaml:"uses,omitempty" validate:"required"`
}

type StepRun struct {
	Step

	// Runs command-line programs using the operating system's shell. If you do not provide a name, the step name will default to the text specified in the run command.
	// Commands run using non-login shells by default. You can choose a different shell and customize the shell used to run commands.
	// For more information, see https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell.
	// Each run keyword represents a new process and shell in the virtual environment. When you provide multi-line commands, each line runs in the same shell.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsrun
	Run Evaluable[string] `json:"run,omitempty" yaml:"run,omitempty" validate:"required"`

	Shell Shell `json:"shell,omitempty" yaml:"shell,omitempty"`

	WorkingDir Evaluable[string] `json:"working-directory,omitempty" yaml:"working-directory,omitempty"`
}

// To set custom environment variables, you need to specify the variables in the workflow file.
// You can define environment variables for a step, job, or entire workflow using the jobs.<job_id>.steps[*].env, jobs.<job_id>.env, and env keywords.
// For more information, see https://docs.github.com/en/actions/learn-github-actions/variables
type Env map[string]Evaluable[string]

type With map[string]Evaluable[string]

type JobOutputs map[string]Evaluable[string]

type JobSecrets struct {
	// A pair consisting of a string identifier for the secret and the value of the secret.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsecretssecret_id
	Secrets map[string]Evaluable[string]

	// Use the inherit keyword to pass all the calling workflow's secrets to the called workflow.
	// This includes all secrets the calling workflow has access to, namely organization, repository, and environment secrets.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsecretsinherit
	Inherit bool `json:"inherit,omitempty" yaml:"inherit,omitempty"`
}

// The environment that the job references
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenvironment
type Environment struct {
	// The name of the environment configured in the repo.
	Name Evaluable[string] `json:"name,omitempty" yaml:"name,omitempty"`
	// A deployment URL
	Url Evaluable[string] `json:"url,omitempty" yaml:"url,omitempty"`
}

type Defaults struct {
	Run struct {
		Shell      Shell  `json:"shell,omitempty" yaml:"shell,omitempty"`
		WorkingDir string `json:"working-directory,omitempty" yaml:"working-directory,omitempty"`
	} `json:"run,omitempty" yaml:"run,omitempty"`
}

type PermissionsLevel string

const (
	PermissionsLevelNone  PermissionsLevel = "none"
	PermissionsLevelRead  PermissionsLevel = "read"
	PermissionsLevelWrite PermissionsLevel = "write"
)

// You can modify the default permissions granted to the GITHUB_TOKEN, adding or removing access as required,
// so that you only allow the minimum required access.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions
type Permissions struct {
	Actions            PermissionsLevel `json:"actions,omitempty" yaml:"actions,omitempty"`
	Checks             PermissionsLevel `json:"checks,omitempty" yaml:"checks,omitempty"`
	Contents           PermissionsLevel `json:"contents,omitempty" yaml:"contents,omitempty"`
	Deployments        PermissionsLevel `json:"deployments,omitempty" yaml:"deployments,omitempty"`
	Discussions        PermissionsLevel `json:"discussions,omitempty" yaml:"discussions,omitempty"`
	IdToken            PermissionsLevel `json:"id-token,omitempty" yaml:"id-token,omitempty"`
	Issues             PermissionsLevel `json:"issues,omitempty" yaml:"issues,omitempty"`
	Packages           PermissionsLevel `json:"packages,omitempty" yaml:"packages,omitempty"`
	Pages              PermissionsLevel `json:"pages,omitempty" yaml:"pages,omitempty"`
	PullRequests       PermissionsLevel `json:"pull-requests,omitempty" yaml:"pull-requests,omitempty"`
	RepositoryProjects PermissionsLevel `json:"repository-projects,omitempty" yaml:"repository-projects,omitempty"`
	SecurityEvents     PermissionsLevel `json:"security-events,omitempty" yaml:"security-events,omitempty"`
	Statuses           PermissionsLevel `json:"statuses,omitempty" yaml:"statuses,omitempty"`
}

// The type of machine to run the job on. The machine can be either a GitHub-hosted runner, or a self-hosted runner.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idruns-on
type RunsOn struct {
	// You can use runs-on to target runner groups, so that the job will execute on any runner that is a member of that group.
	// For more granular control, you can also combine runner groups with labels.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#choosing-runners-in-a-group
	Group string `json:"group,omitempty" yaml:"group,omitempty"`

	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idruns-on
	Labels Evaluable[string] `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// A strategy creates a build matrix for your jobs. You can define different variations of an environment to run each job in.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategy
type Strategy struct {
	Matrix Evaluable[Matrix] `json:"matrix,omitempty" yaml:"matrix,omitempty"`
	// When set to true, GitHub cancels all in-progress jobs if any matrix job fails. Default: true
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategyfail-fast
	FailFast Evaluable[bool] `json:"fail-fast,omitempty" yaml:"fail-fast,omitempty"` // default: true

	// The maximum number of jobs that can run simultaneously when using a matrix job strategy.
	// By default, GitHub will maximize the number of jobs run in parallel depending on the available runners on GitHub-hosted virtual machines.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymax-parallel
	MaxParallel Evaluable[int64] `json:"max-parallel,omitempty" yaml:"max-parallel,omitempty"`
}

type Matrix struct {
}

// Concurrency ensures that only a single job or workflow using the same concurrency group will run at a time.
// A concurrency group can be any string or expression. The expression can use any context except for the secrets context.
// You can also specify concurrency at the workflow level. When a concurrent job or workflow is queued,
// if another job or workflow using the same concurrency group in the repository is in progress,
// the queued job or workflow will be pending. Any previously pending job or workflow in the concurrency group will be canceled.
// To also cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idconcurrency
type Concurrency struct {
	// When a concurrent job or workflow is queued, if another job or workflow using the same concurrency group
	// in the repository is in progress, the queued job or workflow will be pending. Any previously pending job
	// or workflow in the concurrency group will be canceled.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#example-concurrency-groups-1
	Group Evaluable[string] `json:"group,omitempty" yaml:"group,omitempty"`

	// To cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#example-using-concurrency-to-cancel-any-in-progress-job-or-run-1
	CancelInProgress bool `json:"cancel-in-progress,omitempty" yaml:"cancel-in-progress,omitempty"`
}
