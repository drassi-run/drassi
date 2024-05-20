package actions

import "github.com/dungdm93/drassi/core/pkg/model/workflows"

type Action struct {
	// The name of your action. GitHub displays the `name` in the Actions tab to help visually identify actions in each job.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#name
	Name string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty" validate:"required"`

	// The name of the action's author.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#author
	Author string `json:"author,omitempty" yaml:"author,omitempty" mapstructure:"author,omitempty"`

	// A short description of the action.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#description
	Description string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty" validate:"required"`

	// Input parameters allow you to specify data that the action expects to use during runtime.
	// GitHub stores input parameters as environment variables.
	// Input ids with uppercase letters are converted to lowercase during runtime. We recommended using lowercase input ids.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs
	Inputs map[string]workflows.Input `json:"inputs,omitempty" yaml:"inputs,omitempty" mapstructure:"inputs,omitempty"`

	// Output parameters allow you to declare data that an action sets.
	// Actions that run later in a workflow can use the output data set in previously run actions.
	Outputs map[string]workflows.Output `json:"outputs,omitempty" yaml:"outputs,omitempty" mapstructure:"outputs,omitempty"`

	// Specifies whether this is a JavaScript action, a composite action, or a Docker container action and how the action is executed.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#runs
	Runs Runs `json:"runs,omitempty" yaml:"runs,omitempty" mapstructure:"runs,omitempty" validate:"required"`

	// You can use a color and Feather icon to create a badge to personalize and distinguish your action.
	// Badges are shown next to your action name in GitHub Marketplace.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#branding
	Branding Branding `json:"branding,omitempty" yaml:"branding,omitempty" mapstructure:"branding,omitempty"`
}

// You can use a color and Feather icon to create a badge to personalize and distinguish your action.
// Badges are shown next to your action name in GitHub Marketplace.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#branding
type Branding struct {
	// The background color of the badge.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#brandingcolor
	Color string `json:"color,omitempty" yaml:"color,omitempty" mapstructure:"color,omitempty"`

	// The name of the Feather icon to use.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#brandingicon
	Icon string `json:"icon,omitempty" yaml:"icon,omitempty" mapstructure:"icon,omitempty"`
}
