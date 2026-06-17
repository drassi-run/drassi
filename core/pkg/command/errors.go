/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import "errors"

var (
	ErrInvalidCommand       = errors.New("invalid command")
	ErrNotRegisteredCommand = errors.New("not registered command")
)
