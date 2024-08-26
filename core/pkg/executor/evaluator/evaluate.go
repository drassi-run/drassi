package evaluator

import (
	"strings"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/traits"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
)

func Evaluate[R any](env *expression.Env, evaluable workflows.Evaluable[R], v *R) error {
	if evaluable == nil {
		return nil
	}

	u := &unraveler{env: env}
	if res, err := evaluable.Unravel(u); err != nil {
		return err
	} else if err = model.Decode(res, v); err != nil {
		return err
	}

	return nil
}

func Meet(env *expression.Env, condition workflows.Conditional) (bool, error) {
	u := &unraveler{env: env}

	con := string(condition)
	pure := !strings.Contains(con, workflows.OpenExpression)

	if res, err := u.UnravelExpression(con, pure); err != nil {
		return false, err
	} else {
		v := types.NativeToVal(res)

		// See libraries.isTruthy
		b, ok := v.(traits.Logical)
		meet := !ok || b.ToBoolean()
		return meet, nil
	}
}
