/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import "drassi.run/core/pkg/model"

func (c *Container) DecodeMapstructure(input any) (any, error) {
	if image, ok := model.Stringify(input); ok {
		m := map[string]any{"image": image}
		return m, nil
	}
	// process Container normal way
	return input, nil
}
