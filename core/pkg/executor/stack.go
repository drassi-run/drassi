/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

type Stack interface {
	Job() JobExecutor
	Root() StepExecutor
	Leaf() StepExecutor
	Stack() []StepExecutor
}
