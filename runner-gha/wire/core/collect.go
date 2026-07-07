/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_core

import (
	"fmt"

	"drassi.run/core/pkg/executor"
	"go.uber.org/dig"
)

func collectDecorators(scope *dig.Scope) error {
	if err := scope.Provide(collectJobRunDecorators); err != nil {
		return fmt.Errorf("collect JobRunDecorator: %w", err)
	}
	if err := scope.Provide(collectStepRunDecorators); err != nil {
		return fmt.Errorf("collect StepRunDecorator: %w", err)
	}
	if err := scope.Provide(collectActionRunDecorators); err != nil {
		return fmt.Errorf("collect ActionRunDecorator: %w", err)
	}
	return nil
}

type jobRunDecoratorParam struct {
	dig.In

	Timeline executor.JobRunDecorator `name:"timeline"`
	Masker   executor.JobRunDecorator `name:"masker"`
}

func collectJobRunDecorators(p jobRunDecoratorParam) executor.JobRunDecorator {
	return executor.MultiJobRunDecorator{
		p.Timeline,
		p.Masker,
	}
}

type stepRunDecoratorParam struct {
	dig.In

	Timeline executor.StepRunDecorator `name:"timeline"`
	Log      executor.StepRunDecorator `name:"log"`
}

func collectStepRunDecorators(p stepRunDecoratorParam) executor.StepRunDecorator {
	return executor.MultiStepRunDecorator{
		p.Timeline,
		p.Log,
	}
}

type actionRunDecoratorParam struct {
	dig.In

	ConsoleCommand executor.ActionRunDecorator `name:"command"`
}

func collectActionRunDecorators(p actionRunDecoratorParam) executor.ActionRunDecorator {
	return p.ConsoleCommand
}
