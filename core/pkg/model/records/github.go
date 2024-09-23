package records

// The github context contains information about the workflow run and the event that triggered the run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type Github struct {
	Action            string       `json:"action" yaml:"action" actions:"action"`
	ActionPath        string       `json:"action_path" yaml:"action_path" actions:"action_path"`
	ActionRef         string       `json:"action_ref" yaml:"action_ref" actions:"action_ref"`
	ActionRepository  string       `json:"action_repository" yaml:"action_repository" actions:"action_repository"`
	ActionStatus      Result       `json:"action_status" yaml:"action_status" actions:"action_status"`
	Actor             string       `json:"actor" yaml:"actor" actions:"actor"`
	ActorId           string       `json:"actor_id" yaml:"actor_id" actions:"actor_id"`
	ApiUrl            string       `json:"api_url" yaml:"api_url" actions:"api_url"`
	BaseRef           string       `json:"base_ref" yaml:"base_ref" actions:"base_ref"`
	Event             any          `json:"event" yaml:"event" actions:"event"` // TODO data type
	EventName         string       `json:"event_name" yaml:"event_name" actions:"event_name"`
	EventPath         string       `json:"event_path" yaml:"event_path" actions:"event_path"`
	GraphqlUrl        string       `json:"graphql_url" yaml:"graphql_url" actions:"graphql_url"`
	HeadRef           string       `json:"head_ref" yaml:"head_ref" actions:"head_ref"`
	Job               string       `json:"job" yaml:"job" actions:"job"`
	Ref               string       `json:"ref" yaml:"ref" actions:"ref"`
	RefName           string       `json:"ref_name" yaml:"ref_name" actions:"ref_name"`
	RefProtected      bool         `json:"ref_protected" yaml:"ref_protected" actions:"ref_protected"`
	RefType           RefType      `json:"ref_type" yaml:"ref_type" actions:"ref_type"`
	Repository        string       `json:"repository" yaml:"repository" actions:"repository"`
	RepositoryId      string       `json:"repository_id" yaml:"repository_id" actions:"repository_id"`
	RepositoryOwner   string       `json:"repository_owner" yaml:"repository_owner" actions:"repository_owner"`
	RepositoryOwnerId string       `json:"repository_owner_id" yaml:"repository_owner_id" actions:"repository_owner_id"`
	RepositoryUrl     string       `json:"repositoryUrl" yaml:"repositoryUrl" actions:"repositoryUrl"` // naming convention???
	RetentionDays     string       `json:"retention_days" yaml:"retention_days" actions:"retention_days"`
	RunId             string       `json:"run_id" yaml:"run_id" actions:"run_id"`
	RunNumber         string       `json:"run_number" yaml:"run_number" actions:"run_number"`
	RunAttempt        string       `json:"run_attempt" yaml:"run_attempt" actions:"run_attempt"`
	SecretSource      SecretSource `json:"secret_source" yaml:"secret_source" actions:"secret_source"`
	ServerUrl         string       `json:"server_url" yaml:"server_url" actions:"server_url"`
	Sha               string       `json:"sha" yaml:"sha" actions:"sha"`
	Token             string       `json:"token" yaml:"token" actions:"token"`
	TriggeringActor   string       `json:"triggering_actor" yaml:"triggering_actor" actions:"triggering_actor"`
	Workflow          string       `json:"workflow" yaml:"workflow" actions:"workflow"`
	WorkflowRef       string       `json:"workflow_ref" yaml:"workflow_ref" actions:"workflow_ref"`
	WorkflowSha       string       `json:"workflow_sha" yaml:"workflow_sha" actions:"workflow_sha"`
	Workspace         string       `json:"workspace" yaml:"workspace" actions:"workspace"`

	// File commands env
	Path        string `json:"path" yaml:"path" actions:"path"`
	Env         string `json:"env" yaml:"env" actions:"env"`
	Output      string `json:"output" yaml:"output" actions:"output"`
	State       string `json:"state" yaml:"state" actions:"state"`
	StepSummary string `json:"step_summary" yaml:"step_summary" actions:"step_summary"`
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
