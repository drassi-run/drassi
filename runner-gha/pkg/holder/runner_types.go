package holder

import "time"

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
