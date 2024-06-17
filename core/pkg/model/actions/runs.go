package actions

import "github.com/dungdm93/drassi/core/pkg/model/workflows"

type Runs interface {
	isRuns()
}

type JavaScriptRuns struct {
	// The application used to execute the code specified in `main`.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsusing-for-javascript-actions
	Using string `json:"using,omitempty" yaml:"using,omitempty" mapstructure:"using,omitempty" validate:"required"`

	// The file that contains your action code. The application specified in `using` executes this file.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsmain
	Main string `json:"main,omitempty" yaml:"main,omitempty" mapstructure:"main,omitempty" validate:"required"`

	// Allows you to run a script at the start of a job, before the `main:` action begins.
	// For example, you can use `pre:` to run a prerequisite setup script. The application specified with
	// the `using` syntax will execute this file.
	// The `pre:` action always runs by default but you can override this using `pre-if`.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspre
	Pre string `json:"pre,omitempty" yaml:"pre,omitempty" mapstructure:"pre,omitempty"`

	// Allows you to define conditions for the pre: action execution.
	// The pre: action will only run if the conditions in pre-if are met.
	// If not set, then pre-if defaults to always().
	// In pre-if, status check functions evaluate against the job's status, not the action's own status.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspre-if
	PreIf workflows.Conditional `json:"pre-if,omitempty" yaml:"pre-if,omitempty" mapstructure:"pre-if,omitempty"`

	// Allows you to run a script at the end of a job, once the main: action has completed.
	// For example, you can use post: to terminate certain processes or remove unneeded files.
	// The runtime specified with the using syntax will execute this file.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspost
	Post string `json:"post,omitempty" yaml:"post,omitempty" mapstructure:"post,omitempty"`

	// Allows you to define conditions for the post: action execution.
	// The post: action will only run if the conditions in post-if are met.
	// If not set, then post-if defaults to always().
	// In post-if, status check functions evaluate against the job's status, not the action's own status.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspost-if
	PostIf workflows.Conditional `json:"post-if,omitempty" yaml:"post-if,omitempty" mapstructure:"post-if,omitempty"`
}

func (r *JavaScriptRuns) isRuns() {
}

type DockerRuns struct {
	// You must set this value to 'docker'.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsusing-for-docker-container-actions
	Using string `json:"using,omitempty" yaml:"using,omitempty" mapstructure:"using,omitempty" validate:"required"`

	// The Docker image to use as the container to run the action. The value can be the Docker base image name,
	// a local `Dockerfile` in your repository, or a public image in Docker Hub or another registry.
	// To reference a `Dockerfile` local to your repository, use a path relative to your action metadata file.
	// The `docker` application will execute this file.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsimage
	Image string `json:"image,omitempty" yaml:"image,omitempty" mapstructure:"image,omitempty" validate:"required"`

	// Specifies a key/value map of environment variables to set in the container environment.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsenv
	//
	// To set custom environment variables, you need to specify the variables in the workflow file.
	// You can define environment variables for a step, job, or entire workflow using the jobs.<job_id>.steps[*].env, jobs.<job_id>.env, and env keywords.
	// For more information, see https://docs.github.com/en/actions/learn-github-actions/variables
	Env workflows.Evaluable[map[string]string] `json:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`

	// Overrides the Docker `ENTRYPOINT` in the `Dockerfile`, or sets it if one wasn't already specified.
	// Use `entrypoint` when the `Dockerfile` does not specify an `ENTRYPOINT` or you want to override the `ENTRYPOINT` instruction.
	// If you omit `entrypoint`, the commands you specify in the Docker `ENTRYPOINT` instruction will execute.
	// The Docker `ENTRYPOINT instruction has a *shell* form and *exec* form. The Docker `ENTRYPOINT` documentation
	// recommends using the *exec* form of the `ENTRYPOINT` instruction.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsentrypoint
	Entrypoint string `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty" mapstructure:"entrypoint,omitempty"`

	// An array of strings that define the inputs for a Docker container. Inputs can include hardcoded strings.
	// GitHub passes the `args` to the container's `ENTRYPOINT` when the container starts up.
	// The `args` are used in place of the `CMD` instruction in a `Dockerfile`. If you use `CMD` in your `Dockerfile`,
	// use the guidelines ordered by preference:
	// - Document required arguments in the action's README and omit them from the `CMD` instruction.
	// - Use defaults that allow using the action without specifying any `args`.
	// - If the action exposes a `--help` flag, or something similar, use that to make your action self-documenting.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runsargs
	Args []string `json:"args,omitempty" yaml:"args,omitempty" mapstructure:"args,omitempty"`

	// Allows you to run a script before the `entrypoint` action begins. For example, you can use `pre-entrypoint:` to run a prerequisite setup script.
	// GitHub Actions uses `docker run` to launch this action, and runs the script inside a new container that uses the same base image.
	// This means that the runtime state is different from the main `entrypoint` container, and any states you require must be accessed
	// in either the workspace, `HOME`, or as a `STATE_` variable.
	// The `pre-entrypoint:` action always runs by default but you can override this using `pre-if`.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspre-entrypoint
	PreEntrypoint string                `json:"pre-entrypoint,omitempty" yaml:"pre-entrypoint,omitempty" mapstructure:"pre-entrypoint,omitempty"`
	PreIf         workflows.Conditional `json:"pre-if,omitempty" yaml:"pre-if,omitempty" mapstructure:"pre-if,omitempty"`

	// Allows you to run a cleanup script once the `runs.entrypoint` action has completed. GitHub Actions uses `docker run` to launch this action.
	// Because GitHub Actions runs the script inside a new container using the same base image, the runtime state is different from the main `entrypoint` container.
	// You can access any state you need in either the workspace, `HOME`, or as a `STATE_` variable.
	// The `post-entrypoint:` action always runs by default but you can override this using `post-if`.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runspost-entrypoint
	PostEntrypoint string                `json:"post-entrypoint,omitempty" yaml:"post-entrypoint,omitempty" mapstructure:"post-entrypoint,omitempty"`
	PostIf         workflows.Conditional `json:"post-if,omitempty" yaml:"post-if,omitempty" mapstructure:"post-if,omitempty"`
}

func (r *DockerRuns) isRuns() {
}

type CompositeRuns struct {
	// To use a composite run steps action, set this to 'composite'.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs-for-composite-actions
	Using string `json:"using,omitempty" yaml:"using,omitempty" mapstructure:"using,omitempty" validate:"required"`

	// The steps that you plan to run in this action. These can be either run steps or uses steps.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runssteps
	Steps []workflows.Step `json:"steps,omitempty" yaml:"steps,omitempty" mapstructure:"steps,omitempty" validate:"required"`
}

func (r *CompositeRuns) isRuns() {
}
