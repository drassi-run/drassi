/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reporter

import (
	"context"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
)

// https://github.com/actions/runner/blob/main/src/Sdk/DTWebApi/WebApi/Issue.cs
// https://github.com/actions/runner/blob/main/src/Sdk/RSWebApi/Contracts/AnnotationLevel.cs
// https://github.com/actions/runner/blob/main/src/Sdk/RSWebApi/Contracts/IssueExtensions.cs
type IssueType int

const (
	IssueTypeError   IssueType = 1
	IssueTypeWarning IssueType = 2
	IssueTypeNotice  IssueType = 3
)

type Issue struct {
	Type     IssueType         `json:"type,omitempty" yaml:"type,omitempty"`
	Category string            `json:"category,omitempty" yaml:"category,omitempty"`
	Message  string            `json:"message,omitempty" yaml:"message,omitempty"`
	Data     map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

type Reporter interface {
	StartJob(ctx context.Context, je executor.JobExecutor) error
	EndJob(ctx context.Context, je executor.JobExecutor, result *records.Job) error

	StartStep(ctx context.Context, stage executor.Stage, se executor.StepExecutor) error
	EndStep(ctx context.Context, stage executor.Stage, se executor.StepExecutor, result *records.Step) error

	Log(ctx context.Context, msg string) error
	AddIssue(ctx context.Context, issue *Issue) error
	AttachFile(kind, name string, reader io.Reader) error

	Close() error
}
