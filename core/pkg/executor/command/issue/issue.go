/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package issue

import "context"

// https://github.com/actions/runner/blob/main/src/Sdk/DTWebApi/WebApi/Issue.cs
// https://github.com/actions/runner/blob/main/src/Sdk/RSWebApi/Contracts/AnnotationLevel.cs
// https://github.com/actions/runner/blob/main/src/Sdk/RSWebApi/Contracts/IssueExtensions.cs
type Type int

const (
	TypeError   Type = 1
	TypeWarning Type = 2
	TypeNotice  Type = 3
)

type Issue struct {
	Type     Type              `json:"type,omitempty" yaml:"type,omitempty"`
	Category string            `json:"category,omitempty" yaml:"category,omitempty"`
	Message  string            `json:"message,omitempty" yaml:"message,omitempty"`
	Data     map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

type Reporter interface {
	AddIssue(ctx context.Context, issue *Issue) error
}

var Discard Reporter = discard{}

type discard struct{}

func (d discard) AddIssue(context.Context, *Issue) error { return nil }
