/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdtypes

import "context"

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

type Reporter[R any] interface {
	AddIssue(ctx context.Context, res R, issue *Issue) error
}

func Discard[R any]() Reporter[R] {
	return discard[R]{}
}

type discard[R any] struct{}

func (d discard[R]) AddIssue(context.Context, R, *Issue) error { return nil }
