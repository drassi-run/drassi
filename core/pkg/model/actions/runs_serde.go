/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"strings"

	"drassi.run/core/pkg/model"
)

func discriminateRuns(raw jsontext.Value) (Runs, error) {
	var dis struct {
		Using string `json:"using,omitempty"`
	}
	if err := json.Unmarshal(raw, &dis); err != nil {
		return nil, err
	}

	switch u := dis.Using; u {
	case "composite":
		return new(CompositeRuns), nil
	case "docker":
		return new(DockerRuns), nil
	default:
		if strings.HasPrefix(u, "node") {
			return new(NodeRuns), nil
		}
		return nil, fmt.Errorf(`unknown runs with using=%q`, u)
	}
}

func init() {
	model.RegisterUnmarshalInterface(discriminateRuns)
}
