package service

import (
	"time"

	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/gha-runner/pkg/holder"
	"drassi.run/gha-runner/pkg/messages"
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
	SoftSizeLimit int64  `json:"soft_size_limit"`
}

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

// StepLogsMetadataCreate in C#
type metadataStepLogsRequest struct {
	PlanId     string    `json:"workflow_run_backend_id"`     // UUID
	JobId      string    `json:"workflow_job_run_backend_id"` // UUID
	StepId     string    `json:"step_backend_id"`             // UUID
	LineCount  int64     `json:"line_count"`
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

// JobLogsMetadataCreate in C#
type metadataJobLogsRequest struct {
	PlanId     string    `json:"workflow_run_backend_id"`     // UUID
	JobId      string    `json:"workflow_job_run_backend_id"` // UUID
	LineCount  int64     `json:"line_count"`
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

////////////// ResultService: Timeline Records //////////////

// StepsUpdateRequest in C#
type stepsUpdateRequest struct {
	PlanId      string `json:"workflow_run_backend_id"`     // UUID
	JobId       string `json:"workflow_job_run_backend_id"` // UUID
	ChangeOrder int64  `json:"change_order"`
	Steps       []Step `json:"steps"`
}

type Step struct {
	Id          string     `json:"external_id"` // UUID
	Number      int64      `json:"number"`
	Name        string     `json:"name"`
	Status      Status     `json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Conclusion  Conclusion `json:"conclusion"`
}

type Status int

const (
	StatusUnknown    Status = 0
	StatusInProgress Status = 3
	StatusPending    Status = 5
	StatusCompleted  Status = 6
)

type Conclusion int

const (
	ConclusionUnknown   Conclusion = 0
	ConclusionSuccess   Conclusion = 2
	ConclusionFailure   Conclusion = 3
	ConclusionCancelled Conclusion = 4
	ConclusionSkipped   Conclusion = 7
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

////////////// LiveFeed //////////////

// same TimelineRecordLogLine in C# with some extra info
type line struct {
	stepId  string // UUID
	number  int64
	content string
}

// TimelineRecordFeedLinesWrapper in C#
type linesWrapper struct {
	StepUid   string   `json:"step_id"`
	Value     []string `json:"value"`
	Count     int      `json:"count"`
	StartLine int64    `json:"start_line"`
}

////////////// TimelineRecord //////////////

// TimelineRecord in C#
type record struct {
	Id               string                            `json:"id,omitempty"`        // UUID
	TimelineId       string                            `json:"-"`                   // UUID
	ParentId         string                            `json:"parent_id,omitempty"` // UUID
	Type             string                            `json:"type,omitempty"`      // RecordType
	Name             string                            `json:"name,omitempty"`
	StartTime        time.Time                         `json:"start_time,omitempty"`
	FinishTime       time.Time                         `json:"finish_time,omitempty"`
	CurrentOperation string                            `json:"current_operation,omitempty"`
	PercentComplete  int32                             `json:"percent_complete,omitempty"`
	State            holder.TimelineRecordState        `json:"state,omitempty"`
	Result           holder.TaskResult                 `json:"result,omitempty"`
	ResultCode       string                            `json:"result_code,omitempty"`
	ChangeID         int32                             `json:"change_id,omitempty"`
	LastModified     time.Time                         `json:"last_modified,omitempty"`
	WorkerName       string                            `json:"worker_name,omitempty"`
	Order            int32                             `json:"order,omitempty"`
	RefName          string                            `json:"ref_name,omitempty"`
	Log              *TaskLogReference                 `json:"log,omitempty"`
	Details          *TimeLineReference                `json:"details,omitempty"`
	ErrorCount       int                               `json:"error_count,omitempty"`
	WarningCount     int                               `json:"warning_count,omitempty"`
	NoticeCount      int                               `json:"notice_count,omitempty"`
	Issues           []reporter.Issue                  `json:"issues,omitempty"`
	Location         string                            `json:"location,omitempty"`
	Attempt          int32                             `json:"attempt,omitempty"`
	Identifier       string                            `json:"identifier,omitempty"`
	AgentPlatform    string                            `json:"agent_platform,omitempty"`
	PreviousAttempts []TimelineAttempt                 `json:"previous_attempts,omitempty"`
	Variables        map[string]messages.VariableValue `json:"variables,omitempty"`
}

type TaskLogReference struct {
	Id       int32  `json:"id,omitempty"`
	Location string `json:"location,omitempty"`
}

type TimeLineReference struct {
	Id       string `json:"id,omitempty"` // UUID
	ChangeId int32  `json:"change_id,omitempty"`
	Location string `json:"location,omitempty"`
}

type TimelineAttempt struct {
	Identifier string `json:"identifier,omitempty"`
	Attempt    int32  `json:"attempt,omitempty"`
	TimelineId string `json:"timeline_id,omitempty"` // UUID
	RecordId   string `json:"record_id,omitempty"`   // UUID
}

// VssJsonCollectionWrapper in C#
type recordsWrapper struct {
	Count int64     `json:"count,omitempty"`
	Value []*record `json:"value,omitempty"`
}
