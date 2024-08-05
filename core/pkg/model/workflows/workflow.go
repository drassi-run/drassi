package workflows

// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions
type Workflow struct {
	// The name of your workflow. GitHub displays the names of your workflows on your repository's actions page.
	// If you omit this field, GitHub sets the name to the workflow's filename.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#name
	Name string `json:"name,omitempty" yaml:"name,omitempty" actions:"name,omitempty"`

	// The name for workflow runs generated from the workflow.
	// GitHub displays the workflow run name in the list of workflow runs on your repository's 'Actions' tab.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#run-name
	//
	// Context available: `github`, `inputs`, `vars`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	RunName Evaluable[string] `json:"run-name,omitempty" yaml:"run-name,omitempty" actions:"run-name,omitempty"`

	// The name of the GitHub event that triggers the workflow.
	// You can provide a single event string, array of events, array of event types, or an event configuration map
	// that schedules a workflow or restricts the execution of a workflow to specific files, tags, or branch changes.
	// For a list of available events, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#on
	On On `json:"on,omitempty" yaml:"on,omitempty" actions:"on,omitempty"`

	// You can use permissions to modify the default permissions granted to the GITHUB_TOKEN,
	// adding or removing access as required, so that you only allow the minimum required access.
	// For more information, see https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions
	Permissions Permissions `json:"permissions,omitempty" yaml:"permissions,omitempty" actions:"permissions,omitempty"`

	// A map of environment variables that are available to all jobs and steps in the workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#env
	//
	// To set custom environment variables, you need to specify the variables in the workflow file.
	// You can define environment variables for a step, job, or entire workflow using the jobs.<job_id>.steps[*].env, jobs.<job_id>.env, and env keywords.
	// For more information, see https://docs.github.com/en/actions/learn-github-actions/variables
	//
	// Context available: `github`, `secrets`, `inputs`, `vars`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Env Evaluable[map[string]string] `json:"env,omitempty" yaml:"env,omitempty" actions:"env,omitempty"`

	// A map of default settings that will apply to all jobs in the workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#defaults
	Defaults Evaluable[Defaults] `json:"defaults,omitempty" yaml:"defaults,omitempty" actions:"defaults,omitempty"`

	// Concurrency ensures that only a single job or workflow using the same concurrency group will run at a time.
	// A concurrency group can be any string or expression. The expression can use any context except for the secrets context.
	// You can also specify concurrency at the workflow level.
	// When a concurrent job or workflow is queued, if another job or workflow using the same concurrency group
	// in the repository is in progress, the queued job or workflow will be pending.
	// Any previously pending job or workflow in the concurrency group will be canceled.
	// To also cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#concurrency
	Concurrency Concurrency `json:"concurrency,omitempty" yaml:"concurrency,omitempty" actions:"concurrency,omitempty"`

	// A workflow run is made up of one or more jobs. Jobs run in parallel by default.
	// To run jobs sequentially, you can define dependencies on other jobs using the jobs.<job_id>.needs keyword.
	// Each job runs in a fresh instance of the virtual environment specified by runs-on.
	// You can run an unlimited number of jobs as long as you are within the workflow usage limits.
	// For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#usage-limits.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobs
	Jobs map[string]Job `json:"jobs,omitempty" yaml:"jobs,omitempty" actions:"jobs,omitempty"`
}

// You can modify the default permissions granted to the GITHUB_TOKEN, adding or removing access as required,
// so that you only allow the minimum required access.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions
type Permissions struct {
	Actions            PermissionsLevel `json:"actions,omitempty" yaml:"actions,omitempty" actions:"actions,omitempty"`
	Checks             PermissionsLevel `json:"checks,omitempty" yaml:"checks,omitempty" actions:"checks,omitempty"`
	Contents           PermissionsLevel `json:"contents,omitempty" yaml:"contents,omitempty" actions:"contents,omitempty"`
	Deployments        PermissionsLevel `json:"deployments,omitempty" yaml:"deployments,omitempty" actions:"deployments,omitempty"`
	Discussions        PermissionsLevel `json:"discussions,omitempty" yaml:"discussions,omitempty" actions:"discussions,omitempty"`
	IdToken            PermissionsLevel `json:"id-token,omitempty" yaml:"id-token,omitempty" actions:"id-token,omitempty"`
	Issues             PermissionsLevel `json:"issues,omitempty" yaml:"issues,omitempty" actions:"issues,omitempty"`
	Packages           PermissionsLevel `json:"packages,omitempty" yaml:"packages,omitempty" actions:"packages,omitempty"`
	Pages              PermissionsLevel `json:"pages,omitempty" yaml:"pages,omitempty" actions:"pages,omitempty"`
	PullRequests       PermissionsLevel `json:"pull-requests,omitempty" yaml:"pull-requests,omitempty" actions:"pull-requests,omitempty"`
	RepositoryProjects PermissionsLevel `json:"repository-projects,omitempty" yaml:"repository-projects,omitempty" actions:"repository-projects,omitempty"`
	SecurityEvents     PermissionsLevel `json:"security-events,omitempty" yaml:"security-events,omitempty" actions:"security-events,omitempty"`
	Statuses           PermissionsLevel `json:"statuses,omitempty" yaml:"statuses,omitempty" actions:"statuses,omitempty"`
}

type PermissionsLevel string

const (
	PermissionsLevelNone  PermissionsLevel = "none"
	PermissionsLevelRead  PermissionsLevel = "read"
	PermissionsLevelWrite PermissionsLevel = "write"
)

type Defaults struct {
	// Context available:
	// - in workflow level: N/A (not an expression)
	// - in job level: `github`, `needs`, `strategy`, `matrix`, `env`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Run struct {
		Shell      string `json:"shell,omitempty" yaml:"shell,omitempty" actions:"shell,omitempty"`
		WorkingDir string `json:"working-directory,omitempty" yaml:"working-directory,omitempty" actions:"working-directory,omitempty"`
	} `json:"run,omitempty" yaml:"run,omitempty" actions:"run,omitempty"`
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
	//
	// Context available:
	// - in workflow level: `github`, `inputs`, `vars`
	// - in job level: `github`, `needs`, `strategy`, `matrix`, `inputs`, `vars`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Group Evaluable[string] `json:"group,omitempty" yaml:"group,omitempty" actions:"group,omitempty"`

	// To cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#example-using-concurrency-to-cancel-any-in-progress-job-or-run-1
	CancelInProgress bool `json:"cancel-in-progress,omitempty" yaml:"cancel-in-progress,omitempty" actions:"cancel-in-progress,omitempty"`
}
