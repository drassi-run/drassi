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
	return &giteaListener{reporter: reporter}
}

type giteaListener struct {
	executor.NoopJobListener
	executor.NoopStepListener

	reporter *Reporter
}

func (l *giteaListener) OnRunJob(exec executor.JobExecutor) executor.EventHandler {
	return &jobRunEventHandler{exec: exec, reporter: l.reporter}
}

func (l *giteaListener) OnRunStep(exec executor.StepExecutor, stage executor.Stage) executor.EventHandler {
	return &stepRunEventHandler{exec: exec, stage: stage, reporter: l.reporter}
}

type jobRunEventHandler struct {
	exec     executor.JobExecutor
	reporter *Reporter
}

func (h *jobRunEventHandler) Begin(context.Context) error {
	return h.reporter.StartJob(h.exec.JobSpec())
}

func (h *jobRunEventHandler) End(error) error {
	state := h.exec.State()
	return h.reporter.EndJob(h.exec.JobSpec(), state)
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
