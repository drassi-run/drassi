/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"fmt"
	"reflect"

	"drassi.run/core/pkg/model"
)

var typeStep = reflect.TypeFor[Step]()

func DecodeStepHook(from reflect.Value, to reflect.Value) (any, error) {
	if !from.IsValid() || !to.Type().Implements(typeStep) {
		return valueOf(from), nil
	}

	f := from.Interface()
	m, ok := model.ObjectStringify(f)
	if !ok {
		return f, nil
	}

	_, containsRun := m["run"]
	_, containsUses := m["uses"]
	if containsRun == containsUses {
		return nil, fmt.Errorf("map MUST be contains either `run` or `uses`")
	}

	t := to.Interface()
	if containsRun {
		if t == nil {
			to.Set(reflect.ValueOf(&RunActionStep{}))
		} else if _, ok := t.(*RunActionStep); !ok {
			return nil, fmt.Errorf("map contains `run` CAN'T be decode to %T", t)
		}
	}
	if containsUses {
		if t == nil {
			to.Set(reflect.ValueOf(&UsesActionStep{}))
		} else if _, ok := t.(*UsesActionStep); !ok {
			return nil, fmt.Errorf("map contains `uses` CAN'T be decode to %T", t)
		}
	}
	return m, nil
}

func init() {
	model.RegisterDecodeHook(DecodeStepHook)
}
