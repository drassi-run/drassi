package workflows

// A string identifier to associate with the input. The value of <input_id> is a map of the input's metadata.
// The <input_id> must be a unique identifier within the inputs object.
// The <input_id> must start with a letter or _ and contain only alphanumeric characters, -, or _.
// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputsinput_id
type Input struct {
	// Required A string description of the input parameter.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputsinput_iddescription
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// A string shown to users using the deprecated input.
	DeprecationMessage string `json:"deprecationMessage,omitempty" yaml:"deprecationMessage,omitempty"`

	// A boolean to indicate whether the action requires the input parameter. Set to true when the parameter is required.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputsinput_idrequired
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`

	// A string representing the default value. The default value is used when an input parameter isn't specified in a workflow file.
	// https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputsinput_iddefault
	//
	// Context available: `github`, `inputs`, `vars`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Default Evaluable[any] `json:"default,omitempty" yaml:"default,omitempty"` // TODO type args

	// The value of this parameter is a string specifying the data type of the input.
	// This must be one of: boolean, choice, number, environment or string.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_dispatchinputsinput_idtype
	Type InputType `json:"type,omitempty" yaml:"type,omitempty"`

	// The options of the dropdown list, if the type is a choice.
	// https://github.blog/changelog/2021-11-10-github-actions-input-types-for-manual-workflows/
	Options []string `json:"options,omitempty" yaml:"options,omitempty"`
}

type InputType string

const (
	InputTypeString      InputType = "string"
	InputTypeChoice      InputType = "choice"
	InputTypeBoolean     InputType = "boolean"
	InputTypeNumber      InputType = "number"
	InputTypeEnvironment InputType = "environment"
)

type Output struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Context available: `github`, `jobs`, `vars`, `inputs`
	// https://docs.github.com/en/actions/learn-github-actions/contexts#context-availability
	Value Evaluable[string] `json:"value,omitempty" yaml:"value,omitempty"`
}

// A string identifier to associate with the secret.
// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecretssecret_id
type Secret struct {
	// Required A string description of the secret parameter.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// A boolean specifying whether the secret must be supplied.
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#onworkflow_callsecretssecret_idrequired
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`
}
