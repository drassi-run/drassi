/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package gha

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
	HeaderActionsActivityID = "ActivityId"
	HeaderGitHubRequestID   = "X-GitHub-Request-Id"
)

func ParseActionsErrorFromResponse(response *http.Response) error {
	actionsError := ActionsError{
		ActivityID: response.Header.Get(HeaderActionsActivityID),
		StatusCode: response.StatusCode,
	}

	if response.ContentLength == 0 {
		actionsError.Err = errors.New("unknown exception")
		return &actionsError
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		actionsError.Err = err
		return &actionsError
	}

	body = trimByteOrderMark(body)
	contentType, ok := response.Header["Content-Type"]
	if ok && len(contentType) > 0 && strings.Contains(contentType[0], "text/plain") {
		actionsError.Err = errors.New(string(body))
		return &actionsError
	}

	var exception ActionsExceptionError
	if err := json.Unmarshal(body, &exception); err != nil {
		actionsError.Err = err
	} else {
		actionsError.Err = &exception
	}
	return &actionsError
}

// Returns slice of body without utf-8 byte order mark.
// If BOM does not exist body is returned unchanged.
func trimByteOrderMark(body []byte) []byte {
	return bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
}

type ActionsError struct {
	ActivityID string
	StatusCode int
	Err        error
}

func (e *ActionsError) Error() string {
	return fmt.Sprintf("actions error: StatusCode %d, AcivityId %q: %v", e.StatusCode, e.ActivityID, e.Err)
}

func (e *ActionsError) Unwrap() error {
	return e.Err
}

type ActionsExceptionError struct {
	ExceptionName string `json:"typeName,omitempty"`
	Message       string `json:"message,omitempty"`
}

func (e *ActionsExceptionError) Error() string {
	return fmt.Sprintf("%s: %s", e.ExceptionName, e.Message)
}
