/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package evaluator

import (
	"encoding/json/v2"
	"strings"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/types/traits"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/actions"
	"drassi.run/core/pkg/model/workflows"
)

func Evaluate[R any](env expression.Env, evaluable workflows.Evaluable[R], v *R) error {
	if evaluable == nil {
		return nil
	}

	u := &unraveler{env: env}
	if res, err := evaluable.Unravel(u); err != nil {
		return err
	} else if data, err := json.Marshal(res); err != nil {
		return err
	} else {
		un := json.JoinUnmarshalers(
			actions.JsonUnmarshalers(), // actions Unmarshalers already include workflows one
			model.WeakUnmarshalers(),
		)
		return json.Unmarshal(data, v, json.WithUnmarshalers(un))
	}
}

// Meet evaluate [workflows.Conditional] in [expression.Env]
//   - When condition is empty, default to "success()"
//   - When a status function is not referenced, refined "CONDITION" as "success() && <CONDITION>"
//
// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ObjectTemplating/PipelineTemplateConverter.cs#L591
func Meet(env expression.Env, condition workflows.Conditional) (bool, error) {
	con := string(condition)
	if con == "" {
		con = "success()"
	}

	pure := !strings.Contains(con, workflows.OpenExpression)
	node, err := env.Parse(con, pure)
	if err != nil {
		return false, err
	} else {
		r := new(refiner)
		node = r.Refine(node)
	}

	prog, err := env.Bind(node)
	if err != nil {
		return false, err
	}

	res, err := prog.Execute()
	if err != nil {
		return false, err
	}

	// See libraries.isTruthy
	b, ok := res.(traits.Logical)
	meet := !ok || b.ToBoolean()
	return meet, nil
}
