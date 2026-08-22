/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
)

////////////// ResultService: Metadata Response for Create(Job/Step)Logs //////////////

type metadataResponse struct {
	Ok bool `json:"ok"`
}

////////////// ResultService: Step Summary //////////////

// GetSignedStepSummaryURLRequest in C#
type signedUrlStepSummaryRequest struct {
	PlanId string `json:"workflow_run_backend_id"`     // UUID
	JobId  string `json:"workflow_job_run_backend_id"` // UUID
	StepId string `json:"step_backend_id"`             // UUID
}

// GetSignedStepSummaryURLResponse in C#
type signedUrlStepSummaryResponse struct {
	Url           string `json:"summary_url"`
	StorageType   string `json:"blob_storage_type"`
	SoftSizeLimit qint64 `json:"soft_size_limit"`
}

func (s *signedUrlStepSummaryResponse) GetUrl() string          { return s.Url }
func (s *signedUrlStepSummaryResponse) GetStorageType() string  { return s.StorageType }
func (s *signedUrlStepSummaryResponse) GetSoftSizeLimit() int64 { return int64(s.SoftSizeLimit) }

// StepSummaryMetadataCreate in C#
type metadataStepSummaryRequest struct {
	PlanId     string    `json:"workflow_run_backend_id"`     // UUID
	JobId      string    `json:"workflow_job_run_backend_id"` // UUID
	StepId     string    `json:"step_backend_id"`             // UUID
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

////////////// ResultService: Step Logs //////////////

// GetSignedStepLogsURLRequest in C#
type signedUrlStepLogsRequest struct {
	PlanId string `json:"workflow_run_backend_id"`     // UUID
	JobId  string `json:"workflow_job_run_backend_id"` // UUID
	StepId string `json:"step_backend_id"`             // UUID
}

// GetSignedStepLogsURLResponse in C#
type signedUrlStepLogsResponse struct {
	Url           string `json:"logs_url"`
	StorageType   string `json:"blob_storage_type"`
	SoftSizeLimit qint64 `json:"soft_size_limit"`
}

func (s *signedUrlStepLogsResponse) GetUrl() string          { return s.Url }
func (s *signedUrlStepLogsResponse) GetStorageType() string  { return s.StorageType }
func (s *signedUrlStepLogsResponse) GetSoftSizeLimit() int64 { return int64(s.SoftSizeLimit) }

// StepLogsMetadataCreate in C#
type metadataStepLogsRequest struct {
	PlanId     string    `json:"workflow_run_backend_id"`     // UUID
	JobId      string    `json:"workflow_job_run_backend_id"` // UUID
	StepId     string    `json:"step_backend_id"`             // UUID
	LineCount  int       `json:"line_count"`
	UploadedAt time.Time `json:"uploaded_at"`
}

////////////// ResultService: Job Logs //////////////

// GetSignedJobLogsURLRequest in C#
type signedUrlJobLogsRequest struct {
	PlanId string `json:"workflow_run_backend_id"`     // UUID
	JobId  string `json:"workflow_job_run_backend_id"` // UUID
}

// GetSignedJobLogsURLResponse in C#
type signedUrlJobLogsResponse struct {
	Url         string `json:"logs_url"`
	StorageType string `json:"blob_storage_type"`
}

func (s *signedUrlJobLogsResponse) GetUrl() string         { return s.Url }
func (s *signedUrlJobLogsResponse) GetStorageType() string { return s.StorageType }

// JobLogsMetadataCreate in C#
type metadataJobLogsRequest struct {
	PlanId     string    `json:"workflow_run_backend_id"`     // UUID
	JobId      string    `json:"workflow_job_run_backend_id"` // UUID
	LineCount  int       `json:"line_count"`
	UploadedAt time.Time `json:"uploaded_at"`
}

////////////// ResultService: Diagnostic Logs //////////////

// GetSignedDiagnosticLogsURLRequest in C#
type signedUrlDiagnosticLogsRequest struct {
	PlanId string `json:"workflow_run_backend_id"`     // UUID
	JobId  string `json:"workflow_job_run_backend_id"` // UUID
}

// GetSignedDiagnosticLogsURLResponse in C#
type signedUrlDiagnosticLogsResponse struct {
	Url         string `json:"diag_logs_url"`
	StorageType string `json:"blob_storage_type"`
}

func (s *signedUrlDiagnosticLogsResponse) GetUrl() string         { return s.Url }
func (s *signedUrlDiagnosticLogsResponse) GetStorageType() string { return s.StorageType }

////////////// ResultService: Timeline Records //////////////

// StepsUpdateRequest in C#
type stepsUpdateRequest struct {
	PlanId      string  `json:"workflow_run_backend_id"`     // UUID
	JobId       string  `json:"workflow_job_run_backend_id"` // UUID
	ChangeOrder int     `json:"change_order"`
	Steps       []*step `json:"steps"`
}

type step struct {
	Id          string         `json:"external_id"` // UUID
	Number      int            `json:"number"`
	Name        string         `json:"name"`
	Status      stepStatus     `json:"status"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	Conclusion  stepConclusion `json:"conclusion"`
}

type stepStatus int

const (
	stepStatusUnknown    stepStatus = 0
	stepStatusInProgress stepStatus = 3
	stepStatusPending    stepStatus = 5
	stepStatusCompleted  stepStatus = 6
)

type stepConclusion int

const (
	stepConclusionUnknown   stepConclusion = 0
	stepConclusionSuccess   stepConclusion = 2
	stepConclusionFailure   stepConclusion = 3
	stepConclusionCancelled stepConclusion = 4
	stepConclusionSkipped   stepConclusion = 7
)

////////////// JobService: Diagnostic Logs //////////////

// TaskLog in C#
type taskLog struct {
	Id            string    `json:"id"` // UUID
	Location      string    `json:"location"`
	IndexLocation string    `json:"index_location"`
	Path          string    `json:"path"`
	LineCount     int64     `json:"line_count"`
	CreatedOn     time.Time `json:"created_on"`
	LastChangedOn time.Time `json:"last_changed_on"`
}

////////////// TimelineRecord //////////////

// TimelineRecord in C#
type record struct {
	Id               string                       `json:"id,omitempty"`        // UUID
	TimelineId       string                       `json:"-"`                   // UUID
	ParentId         string                       `json:"parent_id,omitempty"` // UUID
	Type             string                       `json:"type,omitempty"`      // RecordType
	Name             string                       `json:"name,omitempty"`
	StartTime        *time.Time                   `json:"start_time,omitempty"`
	FinishTime       *time.Time                   `json:"finish_time,omitempty"`
	CurrentOperation string                       `json:"current_operation,omitempty"`
	PercentComplete  int                          `json:"percent_complete,omitempty"`
	State            timeline.State               `json:"state,omitempty"`
	Result           timeline.Result              `json:"result,omitempty"`
	ResultCode       string                       `json:"result_code,omitempty"`
	ChangeID         int                          `json:"change_id,omitempty"`
	LastModified     time.Time                    `json:"last_modified,omitempty"`
	WorkerName       string                       `json:"worker_name,omitempty"`
	Order            int                          `json:"order,omitempty"`
	RefName          string                       `json:"ref_name,omitempty"`
	Log              *taskLogReference            `json:"log,omitempty"`
	Details          *timelineReference           `json:"details,omitempty"`
	ErrorCount       int                          `json:"error_count,omitempty"`
	WarningCount     int                          `json:"warning_count,omitempty"`
	NoticeCount      int                          `json:"notice_count,omitempty"`
	Issues           []*cmdtypes.Issue            `json:"issues,omitempty"`
	Location         string                       `json:"location,omitempty"`
	Attempt          int                          `json:"attempt,omitempty"`
	Identifier       string                       `json:"identifier,omitempty"`
	AgentPlatform    string                       `json:"agent_platform,omitempty"`
	PreviousAttempts []attempt                    `json:"previous_attempts,omitempty"`
	Variables        map[string]messages.Variable `json:"variables,omitempty"`
}

type taskLogReference struct {
	Id       int32  `json:"id,omitempty"`
	Location string `json:"location,omitempty"`
}

type timelineReference struct {
	Id       string `json:"id,omitempty"` // UUID
	ChangeId int32  `json:"change_id,omitempty"`
	Location string `json:"location,omitempty"`
}

type attempt struct {
	Identifier string `json:"identifier,omitempty"`
	Attempt    int32  `json:"attempt,omitempty"`
	TimelineId string `json:"timeline_id,omitempty"` // UUID
	RecordId   string `json:"record_id,omitempty"`   // UUID
}

////////////// booleans/numbers in quotes //////////////
// Support booleans/numbers represent as string returned by GitHub server.
// In actions/runner (C#) code, it converted automatically by Json.NET (Newtonsoft.Json)

type qint64 int64

func (i *qint64) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	tok, err := d.ReadToken()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}
		return err
	}

	switch tok.Kind() {
	case jsontext.KindNumber:
		n, err := tok.Int()
		if err != nil {
			return err
		}
		*i = qint64(n)

	case jsontext.KindString:
		s := tok.String()
		if val, err := strconv.ParseInt(s, 10, 64); err != nil {
			return fmt.Errorf("parse string %q as int64: %w", s, err)
		} else {
			*i = qint64(val)
		}

	default:
		return fmt.Errorf("unexpected token kind %q", tok.Kind())
	}
	return nil
}
