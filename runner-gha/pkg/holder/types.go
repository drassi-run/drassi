package holder

import (
	"context"
	"time"

	"drassi.run/core/pkg/model/records"
	"drassi.run/gha-runner/pkg/messages"
)

type Lease interface {
	GetMessage() *messages.PipelineAgentJobRequest
	Renew(ctx context.Context)
	Complete(ctx context.Context, result records.Result) error
}

////////////// RunService: request & response //////////////

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
	PlanID         string                            `json:"planId,omitempty"`
	JobID          string                            `json:"jobId,omitempty"`
	Conclusion     string                            `json:"conclusion,omitempty"`
	Outputs        map[string]messages.VariableValue `json:"outputs,omitempty"`
	StepResults    []StepResult                      `json:"stepResults,omitempty"`
	Annotations    []Annotation                      `json:"annotations,omitempty"`
	Telemetry      []Telemetry                       `json:"telemetry,omitempty"`
	EnvironmentUrl string                            `json:"environmentUrl,omitempty"`
	BillingOwnerId string                            `json:"billingOwnerId,omitempty"`
}

////////////// RunnerService: request //////////////

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskAgentJobRequest.cs
type runnerJobRequest struct {
	RequestId              int64         `json:"request_id,omitempty"`
	QueueTime              time.Time     `json:"queue_time,omitempty"`
	AssignTime             time.Time     `json:"assign_time,omitempty"`
	ReceiveTime            time.Time     `json:"receive_time,omitempty"`
	FinishTime             time.Time     `json:"finish_time,omitempty"`
	Result                 TaskResult    `json:"result,omitempty"`
	LockedUntil            time.Time     `json:"locked_until,omitempty"`
	LockToken              string        `json:"lock_token,omitempty"`    // UUID
	ServiceOwner           string        `json:"service_owner,omitempty"` // UUID
	HostId                 string        `json:"host_id,omitempty"`       // UUID
	ScopeId                string        `json:"scope_id,omitempty"`      // UUID
	PlanType               string        `json:"plan_type,omitempty"`
	PlanId                 string        `json:"plan_id,omitempty"` // UUID
	PlanGroup              string        `json:"plan_group,omitempty"`
	QueueId                int           `json:"queue_id,omitempty"`
	PoolId                 int           `json:"pool_id,omitempty"`
	JobId                  string        `json:"job_id,omitempty"` // UUID
	JobName                string        `json:"job_name,omitempty"`
	ExpectedDuration       time.Duration `json:"expected_duration,omitempty"`
	OrchestrationId        string        `json:"orchestration_id,omitempty"`
	MatchesAllAgentsInPool bool          `json:"matches_all_agents_in_pool,omitempty"`
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
