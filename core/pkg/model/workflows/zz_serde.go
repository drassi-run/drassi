/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import "encoding/json/v2"

var unmarshalers []*json.Unmarshalers

func JsonUnmarshalers() *json.Unmarshalers {
	return json.JoinUnmarshalers(unmarshalers...)
}
