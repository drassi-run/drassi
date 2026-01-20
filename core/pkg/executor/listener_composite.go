/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"errors"

	"go.uber.org/dig"
)

func NewCompositeJobListener(listeners ...JobListener) JobListener {
	return &compositeJobListener{listeners: listeners}
}

type compositeJobListener struct {
	listeners []JobListener
}

func (l *compositeJobListener) OnInitializeJob(exec JobExecutor, scope *dig.Scope) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener JobListener) EventHandler {
		return listener.OnInitializeJob(exec, scope)
	})
}

func (l *compositeJobListener) OnRunJob(exec JobExecutor) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener JobListener) EventHandler {
		return listener.OnRunJob(exec)
	})
}

func (l *compositeJobListener) OnRunStage(exec JobExecutor, stage Stage) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener JobListener) EventHandler {
		return listener.OnRunStage(exec, stage)
	})
}

func (l *compositeJobListener) OnFinalizeJob(exec JobExecutor) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener JobListener) EventHandler {
		return listener.OnFinalizeJob(exec)
	})
}

func NewCompositeStepListener(listeners ...StepListener) StepListener {
	return &compositeStepListener{listeners: listeners}
}

type compositeStepListener struct {
	listeners []StepListener
}

func (l *compositeStepListener) OnInitializeStep(exec StepExecutor, scope *dig.Scope) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener StepListener) EventHandler {
		return listener.OnInitializeStep(exec, scope)
	})
}

func (l *compositeStepListener) OnRunStep(exec StepExecutor, stage Stage) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener StepListener) EventHandler {
		return listener.OnRunStep(exec, stage)
	})
}

func (l *compositeStepListener) OnRunTask(exec StepExecutor, task *ActionRun) EventHandler {
	return createCompositeEventHandler(l.listeners, func(listener StepListener) EventHandler {
		return listener.OnRunTask(exec, task)
	})
}

func createCompositeEventHandler[L any](listeners []L, create func(L) EventHandler) EventHandler {
	handlers := make([]EventHandler, 0)
	for _, l := range listeners {
		if h := create(l); h != nil {
			handlers = append(handlers, h)
		}
	}

	switch len(handlers) {
	case 0:
		return nil
	case 1:
		return handlers[0]
	default:
		return NewCompositeEventHandler(handlers...)
	}
}

func NewCompositeEventHandler(handlers ...EventHandler) EventHandler {
	return &compositeEventHandler{handlers: handlers, idx: 0}
}

type compositeEventHandler struct {
	handlers []EventHandler
	idx      int
}

func (eh *compositeEventHandler) Begin(ctx context.Context) error {
	for i, h := range eh.handlers {
		if err := h.Begin(ctx); err != nil {
			return err
		}
		eh.idx = i + 1
	}
	return nil
}

func (eh *compositeEventHandler) End(err error) error {
	var errs []error
	for i := eh.idx - 1; i >= 0; i-- { // for backward
		h := eh.handlers[i]
		errs = append(errs, h.End(err))
	}
	return errors.Join(errs...)
}

func end(eh EventHandler, err *error) {
	if err != nil && *err != nil {
		if ex := eh.End(*err); ex != nil {
			ex = errors.Join(*err, ex)
			*err = ex
		}
	} else {
		if ex := eh.End(nil); ex != nil {
			*err = ex
		}
	}
}
