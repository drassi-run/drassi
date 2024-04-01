package model

// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions
type Workflow struct {
	// The name of your workflow. GitHub displays the names of your workflows on your repository's actions page.
	// If you omit this field, GitHub sets the name to the workflow's filename.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// The name for workflow runs generated from the workflow.
	// GitHub displays the workflow run name in the list of workflow runs on your repository's 'Actions' tab.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#run-name
	RunName Evaluable[string] `json:"run-name,omitempty" yaml:"run-name,omitempty"`

	// The name of the GitHub event that triggers the workflow.
	// You can provide a single event string, array of events, array of event types, or an event configuration map
	// that schedules a workflow or restricts the execution of a workflow to specific files, tags, or branch changes.
	// For a list of available events, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#on
	On On `json:"on,omitempty" yaml:"on,omitempty"`

	// You can use permissions to modify the default permissions granted to the GITHUB_TOKEN,
	// adding or removing access as required, so that you only allow the minimum required access.
	// For more information, see https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions
	Permissions Permissions `json:"permissions,omitempty" yaml:"permissions,omitempty"`

	// A map of environment variables that are available to all jobs and steps in the workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#env
	Env Env `json:"env,omitempty" yaml:"env,omitempty"`

	// A map of default settings that will apply to all jobs in the workflow.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#defaults
	Defaults Defaults `json:"defaults,omitempty" yaml:"defaults,omitempty"`

	// Concurrency ensures that only a single job or workflow using the same concurrency group will run at a time.
	// A concurrency group can be any string or expression. The expression can use any context except for the secrets context.
	// You can also specify concurrency at the workflow level.
	// When a concurrent job or workflow is queued, if another job or workflow using the same concurrency group
	// in the repository is in progress, the queued job or workflow will be pending.
	// Any previously pending job or workflow in the concurrency group will be canceled.
	// To also cancel any currently running job or workflow in the same concurrency group, specify cancel-in-progress: true.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#concurrency
	Concurrency Concurrency `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`

	// A workflow run is made up of one or more jobs. Jobs run in parallel by default.
	// To run jobs sequentially, you can define dependencies on other jobs using the jobs.<job_id>.needs keyword.
	// Each job runs in a fresh instance of the virtual environment specified by runs-on.
	// You can run an unlimited number of jobs as long as you are within the workflow usage limits.
	// For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#usage-limits.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobs
	Jobs map[string]Job `json:"jobs,omitempty" yaml:"jobs,omitempty"`
}
