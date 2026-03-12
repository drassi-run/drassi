/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package timeline

import "context"

type Recorder interface {
	Update(ctx context.Context, records ...*Record) error
}
