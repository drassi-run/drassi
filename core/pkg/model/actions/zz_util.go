/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import "reflect"

func valueOf(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}
