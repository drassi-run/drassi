/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package lease

import (
	"net/http"
	"time"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/types"
)

func newClient(url string, hc *http.Client) (*xhttp.Client, error) {
	client, err := xhttp.NewClient(url)
	if err != nil {
		return nil, err
	}

	client = client.WithDefaultErrorHandler(types.ParseActionsError).
		WithDefaultHeader("User-Agent", "gha-runner") // TODO

	if hc != nil {
		client = client.WithHttpClient(hc)
	}
	return client, nil
}

func renewAt(t time.Time) time.Duration {
	d := time.Until(t)
	if d <= 0 {
		return d
	}

	// Renew when 3/4 time pass or 1 minute before expire, whichever later
	d1 := d * 3 / 4
	d2 := d - time.Minute
	if d1 < d2 {
		return d2
	} else {
		return d1
	}
}
