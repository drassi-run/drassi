package holder

import "time"

type acquireJobRequest struct {
	JobMessageId   string `json:"jobMessageId,omitempty"`
	RunnerOS       string `json:"runnerOS,omitempty"`
	BillingOwnerId string `json:"billingOwnerId,omitempty"`
}

type renewJobRequest struct {
	PlanID string `json:"planId,omitempty"`
	JobId  string `json:"jobId,omitempty"`
}

type renewJobResponse struct {
	LockedUntil time.Time `json:"lockedUntil,omitempty"`
}

type completeJobRequest struct {
	PlanID         string              `json:"planId,omitempty"`
	JobID          string              `json:"jobId,omitempty"`
	Conclusion     string              `json:"conclusion,omitempty"`
	Outputs        map[string]Variable `json:"outputs,omitempty"`
	StepResults    []StepResult        `json:"stepResults,omitempty"`
	Annotations    []Annotation        `json:"annotations,omitempty"`
	Telemetry      []Telemetry         `json:"telemetry,omitempty"`
	EnvironmentUrl string              `json:"environmentUrl,omitempty"`
	BillingOwnerId string              `json:"billingOwnerId,omitempty"`
}

type Variable struct {
	Value    string `json:"value,omitempty"`
	IsSecret bool   `json:"isSecret,omitempty"`
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/StepResult.cs
type StepResult struct {
	Uid               string              `json:"external_id,omitempty"`
	Number            int                 `json:"number,omitempty"`
	Name              string              `json:"name,omitempty"`        // e.g "Run actions/checkout@v3"
	ActionName        string              `json:"action_name,omitempty"` // e.g "actions/checkout"
	Ref               string              `json:"ref,omitempty"`
	Type              string              `json:"type,omitempty"`
	Status            TimelineRecordState `json:"status,omitempty"`
	Conclusion        TaskResult          `json:"conclusion,omitempty"`
	StartedAt         time.Time           `json:"started_at,omitempty"`
	CompletedAt       time.Time           `json:"completed_at,omitempty"`
	CompletedLogURL   string              `json:"completed_log_url,omitempty"`
	CompletedLogLines int64               `json:"completed_log_lines,omitempty"`
	Annotations       []Annotation        `json:"annotations,omitempty"`
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
type AnnotationLevel int

const (
	AnnotationLevelUnknown AnnotationLevel = iota
	AnnotationLevelNotice
	AnnotationLevelWarning
	AnnotationLevelFailure
)

func (l AnnotationLevel) String() string {
	switch l {
	case AnnotationLevelNotice:
		return "notice"
	case AnnotationLevelWarning:
		return "warning"
	case AnnotationLevelFailure:
		return "failure"
	default:
		return "unknown"
	}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/RSWebApi/Contracts/Telemetry.cs
type Telemetry struct {
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TimelineRecordState.cs
type TimelineRecordState int

const (
	TimelineRecordStatePending TimelineRecordState = iota
	TimelineRecordStateInProgress
	TimelineRecordStateCompleted
	TimelineRecordStateDelayed
)

func (s TimelineRecordState) String() string {
	switch s {
	case TimelineRecordStatePending:
		return "pending"
	case TimelineRecordStateInProgress:
		return "in_progress"
	case TimelineRecordStateCompleted:
		return "completed"
	case TimelineRecordStateDelayed:
		return "delayed"
	default:
		return "unknown"
	}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskResult.cs
type TaskResult int

const (
	TaskResultSucceeded TaskResult = iota
	TaskResultSucceededWithIssues
	TaskResultFailed
	TaskResultCanceled
	TaskResultSkipped
	TaskResultAbandoned
)

func (r TaskResult) String() string {
	switch r {
	case TaskResultSucceeded:
		return "succeeded"
	case TaskResultSucceededWithIssues:
		return "succeeded (w/ issues)"
	case TaskResultFailed:
		return "failed"
	case TaskResultCanceled:
		return "canceled"
	case TaskResultSkipped:
		return "skipped"
	case TaskResultAbandoned:
		return "abandoned"
	default:
		return "unknown"
	}
}
