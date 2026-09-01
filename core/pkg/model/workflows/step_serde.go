/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"

	"drassi.run/core/pkg/model"
)

func init() {
	u := model.UnmarshalInterface(discriminateStep)
	unmarshalers = append(unmarshalers, u)
}

func discriminateStep(raw jsontext.Value) (Step, error) {
	var dis struct {
		Run  string `json:"run,omitempty"`
		Uses string `json:"uses,omitempty"`
	}
	if err := json.Unmarshal(raw, &dis); err != nil {
		return nil, err
	}

	switch {
	case dis.Run != "" && dis.Uses != "":
		return nil, fmt.Errorf("step MUST be contains either `run` or `uses`")
	case dis.Run != "":
		return new(RunActionStep), nil
	case dis.Uses != "":
		return new(UsesActionStep), nil
	default:
		// both Run and Uses are missing
		return nil, fmt.Errorf("step MUST be contains either `run` or `uses`")
	}
}
