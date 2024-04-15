package workflows

type Step interface {
	isStep()
}

type BaseStep struct {
	// A unique identifier for the step. You can use the id to reference the step in contexts.
	// For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsid
	Id string `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id,omitempty"`

	// A name for your step to display on GitHub.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsname
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `secrets`, `steps`, `inputs`
	// Special functions: `hashFiles`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Name Evaluable[string] `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`

	// You can use the if conditional to prevent a step from running unless a condition is met.
	// You can use any supported context and expression to create a conditional.
	// Expressions in an if conditional do not require the ${{ }} syntax.
	// For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsif
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `steps`, `inputs`
	// Special functions: `always`, `cancelled`, `success`, `failure`, `hashFiles`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	If Conditional `json:"if,omitempty" yaml:"if,omitempty" mapstructure:"if,omitempty"`

	// A map of the input parameters defined by the action. Each input parameter is a key/value pair.
	// Input parameters are set as environment variables. The variable is prefixed with INPUT_ and converted to upper case.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepswith
	With With `json:"with,omitempty" yaml:"with,omitempty" mapstructure:"with,omitempty"`

	// Sets environment variables for steps to use in the virtual environment. You can also set environment variables for the entire workflow or a job.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsenv
	Env Env `json:"env,omitempty" yaml:"env,omitempty" mapstructure:"env,omitempty"`

	// Prevents a job from failing when a step fails. Set to true to allow a job to pass when this step fails.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepscontinue-on-error
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `secrets`, `steps`, `inputs`
	// Special functions: `hashFiles`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	ContinueOnError Evaluable[bool] `json:"continue-on-error,omitempty" yaml:"continue-on-error,omitempty" mapstructure:"continue-on-error,omitempty"`

	// The maximum number of minutes to run the step before killing the process.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `secrets`, `steps`, `inputs`
	// Special functions: `hashFiles`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	TimeoutMinutes Evaluable[int64] `json:"timeout-minutes,omitempty" yaml:"timeout-minutes,omitempty" mapstructure:"timeout-minutes,omitempty"`
}

type UsesStep struct {
	BaseStep `yaml:",inline" mapstructure:",squash"`

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
	Uses string `json:"uses,omitempty" yaml:"uses,omitempty" mapstructure:"uses,omitempty" validate:"required"`
}

func (s *UsesStep) isStep() {
}

type RunStep struct {
	BaseStep `yaml:",inline" mapstructure:",squash"`

	// Runs command-line programs using the operating system's shell. If you do not provide a name, the step name will default to the text specified in the run command.
	// Commands run using non-login shells by default. You can choose a different shell and customize the shell used to run commands.
	// For more information, see https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell.
	// Each run keyword represents a new process and shell in the virtual environment. When you provide multi-line commands, each line runs in the same shell.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsrun
	//
	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `secrets`, `steps`, `inputs`
	// Special functions: `hashFiles`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Run Evaluable[string] `json:"run,omitempty" yaml:"run,omitempty" mapstructure:"run,omitempty" validate:"required"`

	Shell Shell `json:"shell,omitempty" yaml:"shell,omitempty" mapstructure:"shell,omitempty"`

	// Context available: `github`, `needs`, `strategy`, `matrix`, `job`, `runner`, `env`, `vars`, `secrets`, `steps`, `inputs`
	// Special functions: `hashFiles`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	WorkingDir Evaluable[string] `json:"working-directory,omitempty" yaml:"working-directory,omitempty" mapstructure:"working-directory,omitempty"`
}

func (s *RunStep) isStep() {
}
