/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package records

// The github context contains information about the workflow run and the event that triggered the run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type Github struct {
	Action            string       `json:"action"`
	ActionPath        string       `json:"action_path"`
	ActionRef         string       `json:"action_ref"`
	ActionRepository  string       `json:"action_repository"`
	ActionStatus      Result       `json:"action_status"`
	Actor             string       `json:"actor"`
	ActorId           string       `json:"actor_id"`
	ApiUrl            string       `json:"api_url"`
	BaseRef           string       `json:"base_ref"`
	Event             any          `json:"event"` // TODO data type
	EventName         string       `json:"event_name"`
	EventPath         string       `json:"event_path"`
	GraphqlUrl        string       `json:"graphql_url"`
	HeadRef           string       `json:"head_ref"`
	Job               string       `json:"job"`
	Ref               string       `json:"ref"`
	RefName           string       `json:"ref_name"`
	RefProtected      bool         `json:"ref_protected"`
	RefType           RefType      `json:"ref_type"`
	Repository        string       `json:"repository"`
	RepositoryId      string       `json:"repository_id"`
	RepositoryOwner   string       `json:"repository_owner"`
	RepositoryOwnerId string       `json:"repository_owner_id"`
	RepositoryUrl     string       `json:"repositoryUrl"` // naming convention???
	RetentionDays     string       `json:"retention_days"`
	RunId             string       `json:"run_id"`
	RunNumber         string       `json:"run_number"`
	RunAttempt        string       `json:"run_attempt"`
	SecretSource      SecretSource `json:"secret_source"`
	ServerUrl         string       `json:"server_url"`
	Sha               string       `json:"sha"`
	Token             string       `json:"token"`
	TriggeringActor   string       `json:"triggering_actor"`
	Workflow          string       `json:"workflow"`
	WorkflowRef       string       `json:"workflow_ref"`
	WorkflowSha       string       `json:"workflow_sha"`
	Workspace         string       `json:"workspace"`

	// File commands env
	Path        string `json:"path"`
	Env         string `json:"env"`
	Output      string `json:"output"`
	State       string `json:"state"`
	StepSummary string `json:"step_summary"`
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
