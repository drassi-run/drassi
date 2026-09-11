/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

import "drassi.run/core/pkg/container/types"

func init() {
	types.ParseVolume = ParseVolume
	types.ParsePublish = ParsePublish
	types.ParseExpose = ParseExpose
}
