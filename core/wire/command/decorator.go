/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_command

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	cmd "drassi.run/core/pkg/command"
	"drassi.run/core/pkg/command/cmdtypes"
	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/dig"
	xstring "drassi.run/core/util/string"
)

func NewCommandDecorator(fileMgr cmd.FileManager[exec.Milieu]) exec.ActionRunDecorator {
	return &commandDecorator{fileMgr}
}

type commandDecorator struct {
	fileMgr cmd.FileManager[exec.Milieu]
}

func (c *commandDecorator) DecorateActionRun(task *exec.ActionTask) exec.ActionRun {
	run := task.Run
	res := exec.NewMilieu(task.Stage, task.Executor.StepExecutor())
	return func(ctx context.Context) (rec records.Result, err error) {
		if err = c.fileMgr.Initialize(ctx, res); err != nil {
			rec = records.ResultFailure
			return
		}

		rec, err = run(ctx)

		if ex := c.fileMgr.Process(ctx, res); ex != nil {
			scribe.Errorf(ctx, "process file commands fail: %v", ex)
			if rec == records.ResultSuccess {
				rec = records.ResultFailure
			}
			if rec == records.ResultFailure && err == nil {
				err = fmt.Errorf("process file commands fail: %w", ex)
			}
		}
		return
	}
}

// convert "file" property in issue into workspace relative path if possible.
// used by [middleware.DetectProblem] and [cmdhandler.LogMessage]
func refineIssueFileProp() xdig.Decorator[cmdtypes.Reporter[exec.Milieu]] {
	return func(rep cmdtypes.Reporter[exec.Milieu]) cmdtypes.Reporter[exec.Milieu] {
		return &refinePathReporter{Reporter: rep}
	}
}

type refinePathReporter struct {
	cmdtypes.Reporter[exec.Milieu]
}

func (r *refinePathReporter) AddIssue(ctx context.Context, res exec.Milieu, iss *cmdtypes.Issue) error {
	if file := iss.Data["file"]; file != "" {
		file = filepath.Clean(file)

		// translate path if is in container
		if hpt, ok := res.(cmdtypes.HasPathTranslator); ok {
			if pt := hpt.PathTranslator(); pt != nil {
				if f, ok := pt.TranslatePath(file); ok {
					file = f
				}
			}
		}

		// convert file to workspace relative path & inject repo info
		gh := res.Github()
		ws := xstring.EnsureSuffix(gh.Workspace, "/")

		// convert absolute path into workspace relative path if possible
		if f, ok := strings.CutPrefix(file, ws); ok {
			file = f
		}

		// file is in workspace
		if filepath.IsLocal(file) && !strings.HasPrefix(filepath.Dir(file), "~") {
			iss.Data["repo"] = gh.Repository
		}

		iss.Data["file"] = file
	}

	return r.Reporter.AddIssue(ctx, res, iss)
}
