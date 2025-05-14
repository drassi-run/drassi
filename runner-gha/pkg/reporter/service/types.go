package service

import "time"

////////////// ResultService: Metadata Response for Create(Job/Step)Logs //////////////

type metadataResponse struct {
	Ok bool `json:"ok"`
}

////////////// ResultService: Step Summary //////////////

// GetSignedStepSummaryURLRequest in C#
type signedUrlStepSummaryRequest struct {
	PlanUid string `json:"workflow_run_backend_id"`
	JobUid  string `json:"workflow_job_run_backend_id"`
	StepUid string `json:"step_backend_id"`
}

// GetSignedStepSummaryURLResponse in C#
type signedUrlStepSummaryResponse struct {
	Url           string `json:"summary_url"`
	StorageType   string `json:"blob_storage_type"`
	SoftSizeLimit int64  `json:"soft_size_limit"`
}

// StepSummaryMetadataCreate in C#
type metadataStepSummaryRequest struct {
	PlanUid    string    `json:"workflow_run_backend_id"`
	JobUid     string    `json:"workflow_job_run_backend_id"`
	StepUid    string    `json:"step_backend_id"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

////////////// ResultService: Step Logs //////////////

// GetSignedStepLogsURLRequest in C#
type signedUrlStepLogsRequest struct {
	PlanUid string `json:"workflow_run_backend_id"`
	JobUid  string `json:"workflow_job_run_backend_id"`
	StepUid string `json:"step_backend_id"`
}

// GetSignedStepLogsURLResponse in C#
type signedUrlStepLogsResponse struct {
	Url           string `json:"logs_url"`
	StorageType   string `json:"blob_storage_type"`
	SoftSizeLimit int64  `json:"soft_size_limit"`
}

// StepLogsMetadataCreate in C#
type metadataStepLogsRequest struct {
	PlanUid    string    `json:"workflow_run_backend_id"`
	JobUid     string    `json:"workflow_job_run_backend_id"`
	StepUid    string    `json:"step_backend_id"`
	LineCount  int64     `json:"line_count"`
	UploadedAt time.Time `json:"uploaded_at"`
}

////////////// ResultService: Job Logs //////////////

// GetSignedJobLogsURLRequest in C#
type signedUrlJobLogsRequest struct {
	PlanUid string `json:"workflow_run_backend_id"`
	JobUid  string `json:"workflow_job_run_backend_id"`
}

// GetSignedJobLogsURLResponse in C#
type signedUrlJobLogsResponse struct {
	Url         string `json:"logs_url"`
	StorageType string `json:"blob_storage_type"`
}

// JobLogsMetadataCreate in C#
type metadataJobLogsRequest struct {
	PlanUid    string    `json:"workflow_run_backend_id"`
	JobUid     string    `json:"workflow_job_run_backend_id"`
	LineCount  int64     `json:"line_count"`
	UploadedAt time.Time `json:"uploaded_at"`
}

////////////// ResultService: Diagnostic Logs //////////////

// GetSignedDiagnosticLogsURLRequest in C#
type signedUrlDiagnosticLogsRequest struct {
	PlanUid string `json:"workflow_run_backend_id"`
	JobUid  string `json:"workflow_job_run_backend_id"`
}

// GetSignedDiagnosticLogsURLResponse in C#
type signedUrlDiagnosticLogsResponse struct {
	Url         string `json:"diag_logs_url"`
	StorageType string `json:"blob_storage_type"`
}

////////////// TaskService: Diagnostic Logs //////////////

// TaskLog in C#
type taskLog struct {
	Id            string    `json:"id"`
	Location      string    `json:"location"`
	IndexLocation string    `json:"index_location"`
	Path          string    `json:"path"`
	LineCount     int64     `json:"line_count"`
	CreatedOn     time.Time `json:"created_on"`
	LastChangedOn time.Time `json:"last_changed_on"`
}

////////////// LiveFeed //////////////

// same TimelineRecordLogLine in C# with some extra info
type line struct {
	stepUid string
	number  int64
	content string
}

// TimelineRecordFeedLinesWrapper in C#
type liveFeed struct {
	StepUid   string   `json:"step_id"`
	Lines     []string `json:"value"`
	Count     int      `json:"count"`
	StartLine int64    `json:"start_line"`
}
