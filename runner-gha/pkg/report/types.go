/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"time"
)

const StorageAzureBlob = "BLOB_STORAGE_TYPE_AZURE"

type Stat struct {
	Lines int
	Size  int64
}

func NewStat(lines int, size int64) *Stat {
	return &Stat{lines, size}
}

////////////// ResultService: Metadata Response for Create(Job/Step)Logs //////////////

type metadataResponse struct {
	Ok bool `json:"ok"`
}

type signedUrlResponse interface {
	GetUrl() string
	GetStorageType() string
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
	SoftSizeLimit int64  `json:"soft_size_limit"`
}

func (s *signedUrlStepSummaryResponse) GetUrl() string          { return s.Url }
func (s *signedUrlStepSummaryResponse) GetStorageType() string  { return s.StorageType }
func (s *signedUrlStepSummaryResponse) GetSoftSizeLimit() int64 { return s.SoftSizeLimit }

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
	SoftSizeLimit int64  `json:"soft_size_limit"`
}

func (s *signedUrlStepLogsResponse) GetUrl() string          { return s.Url }
func (s *signedUrlStepLogsResponse) GetStorageType() string  { return s.StorageType }
func (s *signedUrlStepLogsResponse) GetSoftSizeLimit() int64 { return s.SoftSizeLimit }

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
