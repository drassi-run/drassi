/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"gopkg.in/yaml.v3"
)

func decodeWorkflow(payload []byte, workflow *workflows.Workflow) error {
	var raw any
	reader := bytes.NewReader(payload)
	if err := yaml.NewDecoder(reader).Decode(&raw); err != nil && err != io.EOF {
		return err
	}

	return model.Decode(raw, workflow)
}

func convertJobRun(wf *workflows.Workflow) (*executor.JobRun, error) {
	if len(wf.Jobs) > 1 {
		return nil, errors.New("multiple jobs found")
	}
	for jobId, job := range wf.Jobs {
		if nj, ok := job.(*workflows.NormalJob); ok {
			jr := executor.ToJobRun(jobId, nj)
			return jr, nil
		}
		return nil, fmt.Errorf("unsupported job type %T", job)
	}
	return nil, fmt.Errorf("empty job")
}

func convertJobNeeds(taskNeeds map[string]*runnerv1.TaskNeed) map[string]*records.Need {
	if len(taskNeeds) == 0 {
		return nil
	}

	needs := make(map[string]*records.Need, len(taskNeeds))
	for k, n := range taskNeeds {
		needs[k] = &records.Need{
			Outputs: n.Outputs,
			Result:  resultMap[n.Result],
		}
	}
	return needs
}

var resultMap = map[runnerv1.Result]records.Result{
	runnerv1.Result_RESULT_UNSPECIFIED: "",
	runnerv1.Result_RESULT_SUCCESS:     records.ResultSuccess,
	runnerv1.Result_RESULT_FAILURE:     records.ResultFailure,
	runnerv1.Result_RESULT_CANCELLED:   records.ResultCancelled,
	runnerv1.Result_RESULT_SKIPPED:     records.ResultSkipped,
}
