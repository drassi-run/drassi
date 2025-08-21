/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reporter

import (
	"context"
	"drassi.run/core/pkg/executor"
)

func NewListener(reporter *Reporter) executor.Listener {
	return &listener{reporter: reporter}
}

type listener struct {
	executor.NoopJobListener
	executor.NoopStepListener

	reporter *Reporter
}

func (l *listener) OnRunJob(exec executor.JobExecutor) executor.EventHandler {
	return &jobRunEventHandler{exec: exec, reporter: l.reporter}
}

func (l *listener) OnRunStep(exec executor.StepExecutor, stage executor.Stage) executor.EventHandler {
	return &stepRunEventHandler{exec: exec, stage: stage, reporter: l.reporter}
}

type jobRunEventHandler struct {
	exec     executor.JobExecutor
	reporter *Reporter
}

func (h *jobRunEventHandler) Begin(context.Context) error {
	return h.reporter.StartJob(h.exec.JobRun())
}

func (h *jobRunEventHandler) End(error) error {
	state := h.exec.State()
	return h.reporter.EndJob(h.exec.JobRun(), state)
}

type stepRunEventHandler struct {
	exec     executor.StepExecutor
	stage    executor.Stage
	reporter *Reporter
}

func (h *stepRunEventHandler) Begin(context.Context) error {
	return h.reporter.StartStep(h.exec.StepRun(), h.stage)
}

func (h *stepRunEventHandler) End(error) error {
	state := h.exec.State()
	return h.reporter.EndStep(h.exec.StepRun(), h.stage, state)
}
