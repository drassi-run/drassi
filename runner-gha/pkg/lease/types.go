/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package lease

import (
	"context"
	"time"

	"drassi.run/core/pkg/executor/command/issue"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
)

type Lease interface {
	GetMessage() *messages.PipelineAgentJobRequest
	Renew(ctx context.Context)
	Complete(ctx context.Context, record *timeline.Record) error
}

////////////// RunService: request & response //////////////

// AcquireJobRequest in C#
type acquireJobRequest struct {
	JobMessageId   string `json:"jobMessageId,omitempty"`
	RunnerOS       string `json:"runnerOS,omitempty"`
	BillingOwnerId string `json:"billingOwnerId,omitempty"`
}

// RenewJobRequest in C#
type renewJobRequest struct {
	PlanId string `json:"planId,omitempty"`
	JobId  string `json:"jobId,omitempty"`
}

// RenewJobResponse in C#
type renewJobResponse struct {
	LockedUntil time.Time `json:"lockedUntil,omitempty"`
}

// CompleteJobRequest in C#
type completeJobRequest struct {
	PlanId         string                       `json:"planId,omitempty"`
	JobId          string                       `json:"jobId,omitempty"`
	Conclusion     timeline.Result              `json:"conclusion,omitempty"`
	Outputs        map[string]messages.Variable `json:"outputs,omitempty"`
	StepResults    []*StepResult                `json:"stepResults,omitempty"`
	Annotations    []*Annotation                `json:"annotations,omitempty"`
	Telemetry      []*Telemetry                 `json:"telemetry,omitempty"`
	EnvironmentUrl string                       `json:"environmentUrl,omitempty"`
	BillingOwnerId string                       `json:"billingOwnerId,omitempty"`
}

////////////// RunnerService: request //////////////

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskAgentJobRequest.cs
type runnerJobRequest struct {
	RequestId              int64           `json:"request_id,omitempty"`
	QueueTime              time.Time       `json:"queue_time,omitempty"`
	AssignTime             time.Time       `json:"assign_time,omitempty"`
	ReceiveTime            time.Time       `json:"receive_time,omitempty"`
	FinishTime             time.Time       `json:"finish_time,omitempty"`
	Result                 timeline.Result `json:"result,omitempty"`
	LockedUntil            time.Time       `json:"locked_until,omitempty"`
	LockToken              string          `json:"lock_token,omitempty"`    // UUID
	ServiceOwner           string          `json:"service_owner,omitempty"` // UUID
	HostId                 string          `json:"host_id,omitempty"`       // UUID
	ScopeId                string          `json:"scope_id,omitempty"`      // UUID
	PlanType               string          `json:"plan_type,omitempty"`
	PlanId                 string          `json:"plan_id,omitempty"` // UUID
	PlanGroup              string          `json:"plan_group,omitempty"`
	QueueId                int             `json:"queue_id,omitempty"`
	PoolId                 int             `json:"pool_id,omitempty"`
	JobId                  string          `json:"job_id,omitempty"` // UUID
	JobName                string          `json:"job_name,omitempty"`
	ExpectedDuration       time.Duration   `json:"expected_duration,omitempty"`
	OrchestrationId        string          `json:"orchestration_id,omitempty"`
	MatchesAllAgentsInPool bool            `json:"matches_all_agents_in_pool,omitempty"`
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/StepResult.cs
type StepResult struct {
	Id                string          `json:"external_id,omitempty"`
	Number            int             `json:"number,omitempty"`
	Name              string          `json:"name,omitempty"`        // e.g "Run actions/checkout@v3"
	ActionName        string          `json:"action_name,omitempty"` // e.g "actions/checkout"
	ActionRef         string          `json:"ref,omitempty"`
	ActionType        string          `json:"type,omitempty"`
	Status            timeline.State  `json:"status,omitempty"`
	Conclusion        timeline.Result `json:"conclusion,omitempty"`
	StartedAt         *time.Time      `json:"started_at,omitempty"`
	CompletedAt       *time.Time      `json:"completed_at,omitempty"`
	CompletedLogURL   string          `json:"completed_log_url,omitempty"`
	CompletedLogLines int64           `json:"completed_log_lines,omitempty"`
	Annotations       []*Annotation   `json:"annotations,omitempty"`
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/Annotation.cs
type Annotation struct {
	Level                 AnnotationLevel `json:"level,omitempty"`
	Message               string          `json:"message,omitempty"`
	Title                 string          `json:"title,omitempty"`
	RawDetails            string          `json:"rawDetails,omitempty"`
	Path                  string          `json:"path,omitempty"`
	IsInfrastructureIssue bool            `json:"isInfrastructureIssue,omitempty"`
	StartLine             int64           `json:"startLine,omitempty"`
	EndLine               int64           `json:"endLine,omitempty"`
	StartColumn           int64           `json:"startColumn,omitempty"`
	EndColumn             int64           `json:"endColumn,omitempty"`
	StepNumber            int64           `json:"stepNumber,omitempty"`
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/AnnotationLevel.cs
type AnnotationLevel string

const (
	AnnotationLevelUnknown AnnotationLevel = ""
	AnnotationLevelNotice  AnnotationLevel = "notice"
	AnnotationLevelWarning AnnotationLevel = "warning"
	AnnotationLevelFailure AnnotationLevel = "failure"
)

func ToAnnotationLevel(t issue.Type) AnnotationLevel {
	switch t {
	case issue.TypeError:
		return AnnotationLevelFailure
	case issue.TypeWarning:
		return AnnotationLevelWarning
	case issue.TypeNotice:
		return AnnotationLevelNotice
	default:
		return AnnotationLevelUnknown
	}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/Telemetry.cs
type Telemetry struct {
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}
