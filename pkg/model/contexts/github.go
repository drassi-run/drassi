package contexts

type Github struct {
	Action            string       `json:"action" yaml:"action"`
	ActionPath        string       `json:"action_path" yaml:"action_path"`
	ActionRef         string       `json:"action_ref" yaml:"action_ref"`
	ActionRepository  string       `json:"action_repository" yaml:"action_repository"`
	ActionStatus      string       `json:"action_status" yaml:"action_status"`
	Actor             string       `json:"actor" yaml:"actor"`
	ActorId           string       `json:"actor_id" yaml:"actor_id"`
	ApiUrl            string       `json:"api_url" yaml:"api_url"`
	BaseRef           string       `json:"base_ref" yaml:"base_ref"`
	Env               string       `json:"env" yaml:"env"`
	Event             any          `json:"event" yaml:"event"` // TODO data type
	EventName         string       `json:"event_name" yaml:"event_name"`
	EventPath         string       `json:"event_path" yaml:"event_path"`
	GraphqlUrl        string       `json:"graphql_url" yaml:"graphql_url"`
	HeadRef           string       `json:"head_ref" yaml:"head_ref"`
	Job               string       `json:"job" yaml:"job"`
	Path              string       `json:"path" yaml:"path"`
	Ref               string       `json:"ref" yaml:"ref"`
	RefName           string       `json:"ref_name" yaml:"ref_name"`
	RefProtected      bool         `json:"ref_protected" yaml:"ref_protected"`
	RefType           RefType      `json:"ref_type" yaml:"ref_type"`
	Repository        string       `json:"repository" yaml:"repository"`
	RepositoryId      string       `json:"repository_id" yaml:"repository_id"`
	RepositoryOwner   string       `json:"repository_owner" yaml:"repository_owner"`
	RepositoryOwnerId string       `json:"repository_owner_id" yaml:"repository_owner_id"`
	RepositoryUrl     string       `json:"repositoryUrl" yaml:"repositoryUrl"` // naming convention???
	RetentionDays     string       `json:"retention_days" yaml:"retention_days"`
	RunId             string       `json:"run_id" yaml:"run_id"`
	RunNumber         string       `json:"run_number" yaml:"run_number"`
	RunAttempt        string       `json:"run_attempt" yaml:"run_attempt"`
	SecretSource      SecretSource `json:"secret_source" yaml:"secret_source"`
	ServerUrl         string       `json:"server_url" yaml:"server_url"`
	Sha               string       `json:"sha" yaml:"sha"`
	Token             string       `json:"token" yaml:"token"`
	TriggeringActor   string       `json:"triggering_actor" yaml:"triggering_actor"`
	Workflow          string       `json:"workflow" yaml:"workflow"`
	WorkflowRef       string       `json:"workflow_ref" yaml:"workflow_ref"`
	WorkflowSha       string       `json:"workflow_sha" yaml:"workflow_sha"`
	Workspace         string       `json:"workspace" yaml:"workspace"`
}

type RefType string

const (
	RefTypeBranch = "branch"
	RefTypeTag    = "tag"
)

type SecretSource string

const (
	SecretSourceNone       = "None"
	SecretSourceActions    = "Actions"
	SecretSourceCodespaces = "Codespaces"
	SecretSourceDependabot = "Dependabot"
)
