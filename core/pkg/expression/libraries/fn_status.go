package libraries

import "drassi.run/core/pkg/model/dossiers"

type JobStatus struct {
	jobInfo *dossiers.Job
}

func (s *JobStatus) Always() bool {
	return true
}

func (s *JobStatus) Success() bool {
	return s.jobInfo.Status == dossiers.ResultSuccess
}

func (s *JobStatus) Failure() bool {
	return s.jobInfo.Status == dossiers.ResultFailure
}

func (s *JobStatus) Cancelled() bool {
	return s.jobInfo.Status == dossiers.ResultCancelled
}

type StepStatus struct {
	gh *dossiers.Github
}

func (s *StepStatus) Always() bool {
	return true
}

func (s *StepStatus) Success() bool {
	return s.gh.ActionStatus == dossiers.ResultSuccess
}

func (s *StepStatus) Failure() bool {
	return s.gh.ActionStatus == dossiers.ResultFailure
}

func (s *StepStatus) Cancelled() bool {
	return s.gh.ActionStatus == dossiers.ResultCancelled
}
