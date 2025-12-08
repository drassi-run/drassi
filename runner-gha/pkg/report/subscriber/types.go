/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package subscriber

import "drassi.run/gha-runner/pkg/log"

type Subscriber interface {
	Run(ch <-chan *log.Event)
	Wait()
}
