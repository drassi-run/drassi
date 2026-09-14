/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package records

import (
	"net/url"
	"path/filepath"
	"strings"

	"drassi.run/core/util/string"
)

// Forge (a.k.a `github`) is the context contains information about the workflow run and the event that triggered the run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type Forge struct {
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

const (
	LabelRepository = "run.drassi.repository"
	LabelReference  = "run.drassi.reference"
	LabelWorkflow   = "run.drassi.workflow"
	LabelJob        = "run.drassi.job"
	LabelAttempt    = "run.drassi.attempt"
	LabelRun        = "run.drassi.run"
)

func (f *Forge) WellKnownLabels() map[string]string {
	repo := f.Repository
	if u, err := url.Parse(f.ServerUrl); err == nil {
		if server := u.Host; server != "" {
			server = strings.ToLower(server)
			server = strings.TrimRight(server, "/")
			repo = server + "/" + repo
		}
	}

	labels := map[string]string{
		LabelRepository: repo,         // e.g: github.com/drassi-run/drassi
		LabelReference:  f.Ref,        // e.g: refs/heads/main
		LabelWorkflow:   f.Workflow,   // e.g: test
		LabelJob:        f.Job,        // e.g: unittests
		LabelAttempt:    f.RunAttempt, // e.g: 1
		LabelRun:        f.RunId,      // e.g: 11208400917
	}
	return labels
}

func (f *Forge) CanonicalName() string {
	repo := xstring.Normalize(f.Repository)
	repo = strings.ToLower(repo)

	workflow := strings.TrimSuffix(f.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(f.Job)
	run := xstring.Normalize(f.RunId)
	attempt := xstring.Normalize(f.RunAttempt)

	name := strings.Join([]string{repo, workflow, job, run, attempt}, "-")
	return name
}

func (f *Forge) StandardPath() string {
	var server string
	if u, err := url.Parse(f.ServerUrl); err == nil {
		server = u.Host
	}
	server = strings.ToLower(server)
	repo := strings.ToLower(f.Repository)

	workflow := strings.TrimSuffix(f.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(f.Job)
	run := xstring.Normalize(f.RunId)
	attempt := xstring.Normalize(f.RunAttempt)

	path := filepath.Join(server, repo, workflow, job, run+"_"+attempt)
	return path
}
