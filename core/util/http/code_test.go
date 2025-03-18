/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utilhttp

import (
	"net/http"
	"testing"
)

func TestStatusType(t *testing.T) {
	code := http.StatusContinue
	if !IsInformational(code) {
		t.Errorf("Status %s is not 1xx: Informational", http.StatusText(code))
	}

	code = http.StatusCreated
	if IsInformational(code) {
		t.Errorf("Status %s is 1xx: Informational", http.StatusText(code))
	}

	code = http.StatusOK
	if !IsSuccess(code) {
		t.Errorf("Status %s is not 2xx: Success", http.StatusText(code))
	}

	code = http.StatusProcessing
	if IsSuccess(code) {
		t.Errorf("Status %s is 2xx: Success", http.StatusText(code))
	}

	code = http.StatusFound
	if !IsRedirection(code) {
		t.Errorf("Status %s is not 3xx: Redirection", http.StatusText(code))
	}

	code = http.StatusNotFound
	if IsRedirection(code) {
		t.Errorf("Status %s is 3xx: Redirection", http.StatusText(code))
	}

	code = http.StatusBadRequest
	if !IsClientError(code) {
		t.Errorf("Status %s is not 4xx: ClientError", http.StatusText(code))
	}

	code = http.StatusGatewayTimeout
	if IsClientError(code) {
		t.Errorf("Status %s is 4xx: ClientError", http.StatusText(code))
	}

	code = http.StatusInternalServerError
	if !IsServerError(code) {
		t.Errorf("Status %s is not 5xx: ServerError", http.StatusText(code))
	}

	code = http.StatusRequestURITooLong
	if IsServerError(code) {
		t.Errorf("Status %s is 5xx: ServerError", http.StatusText(code))
	}
}
