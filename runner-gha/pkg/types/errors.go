/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Header names for request IDs
const (
	HeaderActivityID      = "ActivityId"
	HeaderGitHubRequestID = "X-GitHub-Request-Id"
)

func ParseActionsError(code int, header http.Header, body io.Reader) error {
	actionsError := ActionsError{
		ActivityID: header.Get(HeaderActivityID),
		StatusCode: code,
	}

	data, err := io.ReadAll(body)
	if err != nil {
		actionsError.Err = err
		return &actionsError
	}

	data = bytes.TrimPrefix(data, Utf8BOM)
	contentType := header.Get("Content-Type")
	if contentType == "" {
		return &actionsError
	}

	if strings.HasPrefix(contentType, "text/plain") {
		actionsError.Err = errors.New(string(data))
	} else if strings.HasPrefix(contentType, "application/json") {
		var exception ActionsException
		if err = json.Unmarshal(data, &exception); err != nil {
			actionsError.Err = fmt.Errorf("error unmarshalling actions exception: %w", err)
		} else {
			actionsError.Err = &exception
		}
	}

	return &actionsError
}

type ActionsError struct {
	ActivityID string
	StatusCode int
	Err        error
}

func (e *ActionsError) Error() string {
	return fmt.Sprintf("ActionsError: StatusCode %d, AcivityId %q: %v", e.StatusCode, e.ActivityID, e.Err)
}

func (e *ActionsError) Unwrap() error {
	return e.Err
}

type ActionsException struct {
	ExceptionName string `json:"typeName,omitempty"`
	Message       string `json:"message,omitempty"`
}

func (e *ActionsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ExceptionName, e.Message)
}
