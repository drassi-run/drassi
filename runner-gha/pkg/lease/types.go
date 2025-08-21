/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package lease

import (
	"context"
	"time"

	"drassi.run/core/pkg/executor/support"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
)

type Lease interface {
	GetMessage() *messages.PipelineAgentJobRequest
	Renew(ctx context.Context)
	Complete(ctx context.Context, record *types.Record) error
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
	PlanID string `json:"planId,omitempty"`
	JobId  string `json:"jobId,omitempty"`
}

// RenewJobResponse in C#
type renewJobResponse struct {
	LockedUntil time.Time `json:"lockedUntil,omitempty"`
}

// CompleteJobRequest in C#
type completeJobRequest struct {
	PlanID         string                       `json:"planId,omitempty"`
	JobID          string                       `json:"jobId,omitempty"`
	Conclusion     types.Result                 `json:"conclusion,omitempty"`
	Outputs        map[string]messages.Variable `json:"outputs,omitempty"`
	StepResults    []*StepResult                `json:"stepResults,omitempty"`
	Annotations    []*Annotation                `json:"annotations,omitempty"`
	Telemetry      []*Telemetry                 `json:"telemetry,omitempty"`
	EnvironmentUrl string                       `json:"environmentUrl,omitempty"`
	BillingOwnerId string                       `json:"billingOwnerId,omitempty"`
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/StepResult.cs
type StepResult struct {
	Id                string        `json:"external_id,omitempty"`
	Number            int           `json:"number,omitempty"`
	Name              string        `json:"name,omitempty"`        // e.g "Run actions/checkout@v3"
	ActionName        string        `json:"action_name,omitempty"` // e.g "actions/checkout"
	ActionRef         string        `json:"ref,omitempty"`
	ActionType        string        `json:"type,omitempty"`
	Status            types.State   `json:"status,omitempty"`
	Conclusion        types.Result  `json:"conclusion,omitempty"`
	StartedAt         *time.Time    `json:"started_at,omitempty"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty"`
	CompletedLogURL   string        `json:"completed_log_url,omitempty"`
	CompletedLogLines int64         `json:"completed_log_lines,omitempty"`
	Annotations       []*Annotation `json:"annotations,omitempty"`
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

func ToAnnotationLevel(t support.IssueType) AnnotationLevel {
	switch t {
	case support.IssueTypeError:
		return AnnotationLevelFailure
	case support.IssueTypeWarning:
		return AnnotationLevelWarning
	case support.IssueTypeNotice:
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
