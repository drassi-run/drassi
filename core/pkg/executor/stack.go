/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"io"
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

type Tracker interface {
	Context() context.Context
	StartContext(ctx context.Context) func()

	Env() map[string]string
	ProvideEnv(func() map[string]string)

	AddIssue(ctx context.Context, issue *Issue) error
	AttachFile(ctx context.Context, kind, name string, reader io.Reader) error
}

type Stack interface {
	Job() JobExecutor
	Root() StepExecutor
	Leaf() StepExecutor
	Stack() []StepExecutor
}
