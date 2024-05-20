package contexts

// The github context contains information about the workflow run and the event that triggered the run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type Github struct {
	Action            string       `json:"action" yaml:"action" mapstructure:"action"`
	ActionPath        string       `json:"action_path" yaml:"action_path" mapstructure:"action_path"`
	ActionRef         string       `json:"action_ref" yaml:"action_ref" mapstructure:"action_ref"`
	ActionRepository  string       `json:"action_repository" yaml:"action_repository" mapstructure:"action_repository"`
	ActionStatus      string       `json:"action_status" yaml:"action_status" mapstructure:"action_status"`
	Actor             string       `json:"actor" yaml:"actor" mapstructure:"actor"`
	ActorId           string       `json:"actor_id" yaml:"actor_id" mapstructure:"actor_id"`
	ApiUrl            string       `json:"api_url" yaml:"api_url" mapstructure:"api_url"`
	BaseRef           string       `json:"base_ref" yaml:"base_ref" mapstructure:"base_ref"`
	Env               string       `json:"env" yaml:"env" mapstructure:"env"`
	Event             any          `json:"event" yaml:"event" mapstructure:"event"`
	EventName         string       `json:"event_name" yaml:"event_name" mapstructure:"event_name"`
	EventPath         string       `json:"event_path" yaml:"event_path" mapstructure:"event_path"`
	GraphqlUrl        string       `json:"graphql_url" yaml:"graphql_url" mapstructure:"graphql_url"`
	HeadRef           string       `json:"head_ref" yaml:"head_ref" mapstructure:"head_ref"`
	Job               string       `json:"job" yaml:"job" mapstructure:"job"`
	Path              string       `json:"path" yaml:"path" mapstructure:"path"`
	Ref               string       `json:"ref" yaml:"ref" mapstructure:"ref"`
	RefName           string       `json:"ref_name" yaml:"ref_name" mapstructure:"ref_name"`
	RefProtected      bool         `json:"ref_protected" yaml:"ref_protected" mapstructure:"ref_protected"`
	RefType           RefType      `json:"ref_type" yaml:"ref_type" mapstructure:"ref_type"`
	Repository        string       `json:"repository" yaml:"repository" mapstructure:"repository"`
	RepositoryId      string       `json:"repository_id" yaml:"repository_id" mapstructure:"repository_id"`
	RepositoryOwner   string       `json:"repository_owner" yaml:"repository_owner" mapstructure:"repository_owner"`
	RepositoryOwnerId string       `json:"repository_owner_id" yaml:"repository_owner_id" mapstructure:"repository_owner_id"`
	RepositoryUrl     string       `json:"repositoryUrl" yaml:"repositoryUrl" mapstructure:"repositoryUrl"`
	RetentionDays     string       `json:"retention_days" yaml:"retention_days" mapstructure:"retention_days"`
	RunId             string       `json:"run_id" yaml:"run_id" mapstructure:"run_id"`
	RunNumber         string       `json:"run_number" yaml:"run_number" mapstructure:"run_number"`
	RunAttempt        string       `json:"run_attempt" yaml:"run_attempt" mapstructure:"run_attempt"`
	SecretSource      SecretSource `json:"secret_source" yaml:"secret_source" mapstructure:"secret_source"`
	ServerUrl         string       `json:"server_url" yaml:"server_url" mapstructure:"server_url"`
	Sha               string       `json:"sha" yaml:"sha" mapstructure:"sha"`
	Token             string       `json:"token" yaml:"token" mapstructure:"token"`
	TriggeringActor   string       `json:"triggering_actor" yaml:"triggering_actor" mapstructure:"triggering_actor"`
	Workflow          string       `json:"workflow" yaml:"workflow" mapstructure:"workflow"`
	WorkflowRef       string       `json:"workflow_ref" yaml:"workflow_ref" mapstructure:"workflow_ref"`
	WorkflowSha       string       `json:"workflow_sha" yaml:"workflow_sha" mapstructure:"workflow_sha"`
	Workspace         string       `json:"workspace" yaml:"workspace" mapstructure:"workspace"`
}

type RefType string

const (
	RefTypeBranch RefType = "branch"
	RefTypeTag    RefType = "tag"
)

type SecretSource string

const (
	SecretSourceNone       SecretSource = "None"
	SecretSourceActions    SecretSource = "Actions"
	SecretSourceCodespaces SecretSource = "Codespaces"
	SecretSourceDependabot SecretSource = "Dependabot"
)
