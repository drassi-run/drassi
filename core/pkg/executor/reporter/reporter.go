package reporter

import (
	"io"

	"drassi.run/core/pkg/model/contexts"
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
	Stdin() io.Reader
	Stdout() io.Writer
	Stderr() io.Writer

	StartJob()
	EndJob(result contexts.Result, outputs map[string]string)
	StartStep(stepId string)
	EndStep(stepId string, result contexts.Result)

	AddIssue(issue *Issue) error
	AttachFile(kind, name string, reader io.Reader) error

	Close() error
}
