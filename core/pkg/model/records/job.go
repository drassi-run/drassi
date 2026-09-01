/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package records

// The job context contains information about the currently running job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/JobContext.cs
type JobInfo struct {
	Container *ContainerInfo            `json:"container"`
	Services  map[string]*ContainerInfo `json:"services"`
	Status    Result                    `json:"status"`
}

type ContainerInfo struct {
	Id      string            `json:"id"`
	Network string            `json:"network"`
	Ports   map[string]string `json:"ports"`
}

// The `jobs` context is only available in reusable workflows, and can only be used to set outputs for a reusable workflow.
// https://docs.github.com/en/actions/learn-github-actions/contexts#jobs-context
type JobResult struct {
	Result  Result            `json:"result"`
	Outputs map[string]string `json:"outputs"`
}

// The `steps` context contains information about the steps in the current job that have an `id` specified and have already run.
// https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/StepsContext.cs
type StepResult struct {
	Outputs    map[string]string `json:"outputs"`
	Conclusion Result            `json:"conclusion"`
	Outcome    Result            `json:"outcome"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionResult.cs
type Result string

const (
	ResultSuccess   Result = "success"
	ResultFailure   Result = "failure"
	ResultCancelled Result = "cancelled"
	ResultSkipped   Result = "skipped"
)
