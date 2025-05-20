package types

import (
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/records"
)

type Record struct {
	Uid   string
	Order int

	Object   any
	Children []*Record

	State       State
	Result      Result
	StartedAt   *time.Time
	CompletedAt *time.Time
	Issues      []reporter.Issue
}

// Result equals to TaskResult in C# of GitHub actions/runner
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskResult.cs
type Result string

const (
	ResultUnknown   Result = ""
	ResultSucceeded Result = "succeeded"
	ResultFailed    Result = "failed"
	ResultCanceled  Result = "canceled"
	ResultSkipped   Result = "skipped"
	ResultAbandoned Result = "abandoned"
)

// ToResult converts a records.Result (ActionResult) to a Result (TaskResult).
// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/Util/TaskResultUtil.cs#L62
func ToResult(r records.Result) Result {
	switch r {
	case records.ResultSuccess:
		return ResultSucceeded
	case records.ResultFailure:
		return ResultFailed
	case records.ResultCancelled:
		return ResultCanceled
	case records.ResultSkipped:
		return ResultSkipped
	default:
		return ResultUnknown
	}
}

// State equals to TimelineRecordState in C# of GitHub actions/runner
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TimelineRecordState.cs
type State string

const (
	StateUnknown    State = ""
	StatePending    State = "pending"
	StateInProgress State = "in_progress"
	StateCompleted  State = "completed"
	StateDelayed    State = "delayed"
)

type JobObject struct {
	JobRun         executor.JobRun
	Outputs        map[string]string
	EnvironmentUrl string
}

type StepObject struct {
	StepRun executor.StepRun
	Stage   executor.Stage
}
