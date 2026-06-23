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

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model"
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

func convertJobSpec(wf *workflows.Workflow) (*executor.JobSpec, error) {
	if len(wf.Jobs) > 1 {
		return nil, errors.New("multiple jobs found")
	}
	for jobId, job := range wf.Jobs {
		if nj, ok := job.(*workflows.NormalJob); ok {
			spec := executor.ToJobSpec(jobId, nj)
			return spec, nil
		}
		return nil, fmt.Errorf("unsupported job type %T", job)
	}
	return nil, fmt.Errorf("empty job")
}
