/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"fmt"
	"reflect"
)

func valueOf(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

func castMap[K comparable, V any](input map[K]any) (map[K]V, error) {
	res := make(map[K]V, len(input))
	for k, v := range input {
		if e, ok := v.(V); ok {
			res[k] = e
			continue
		}
		return nil, fmt.Errorf("expected map of %s, got a value %#v type %T", reflect.TypeFor[V]().String(), v, v)
	}
	return res, nil
}
