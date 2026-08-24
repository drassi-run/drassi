/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"drassi.run/core/pkg/command"
	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/problem"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/pkg/stream"
)

type Middleware[R any] func(sink stream.Sink[R]) stream.Sink[R]

func ProcessCommand[R any](consMgr command.ConsoleManager[R]) Middleware[R] {
	return func(sink stream.Sink[R]) stream.Sink[R] {
		return &commandProcessor[R]{
			sink:    sink,
			consMgr: consMgr,
		}
	}
}

type commandProcessor[R any] struct {
	sink    stream.Sink[R]
	consMgr command.ConsoleManager[R]
}

func (mw *commandProcessor[R]) Emit(ctx context.Context, res R, line string) error {
	cmd := mw.consMgr.ParseCommand(line)
	if cmd == nil {
		return mw.sink.Emit(ctx, res, line)
	}

	if err := mw.consMgr.Process(ctx, res, line, cmd); err != nil {
		return fmt.Errorf("process command %q: %w", cmd.Name, err)
	}
	return nil
}

func DetectProblem[R any](factory problem.DetectorFactory, reporter cmdtypes.Reporter[R]) Middleware[R] {
	return func(sink stream.Sink[R]) stream.Sink[R] {
		return &problemDetector[R]{
			sink:     sink,
			reporter: reporter,
			detector: factory.NewDetector(),
		}
	}
}

type problemDetector[R any] struct {
	sink     stream.Sink[R]
	reporter cmdtypes.Reporter[R]
	detector problem.Detector
}

func (mw *problemDetector[R]) Emit(ctx context.Context, res R, line string) (err error) {
	pbl := mw.detect(line)
	if pbl != nil {
		err = mw.report(ctx, res, pbl)
	}

	err2 := mw.sink.Emit(ctx, res, line)
	return errors.Join(err, err2)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (mw *problemDetector[R]) detect(line string) *problem.Problem {
	line = colorCodeRegex.ReplaceAllLiteralString(line, "")
	return mw.detector.Detect(line)
}

func (mw *problemDetector[R]) report(ctx context.Context, res R, pbl *problem.Problem) error {
	iss, err := mw.toIssuer(pbl)
	if err != nil {
		return err
	}

	return mw.reporter.AddIssue(ctx, res, iss)
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (mw *problemDetector[R]) toIssuer(pbl *problem.Problem) (*cmdtypes.Issue, error) {
	if pbl.Message == "" {
		return nil, fmt.Errorf("%s empty message", skippedIssueMsg)
	}
	iss := &cmdtypes.Issue{
		Message: pbl.Message,
		Data:    make(map[string]string),
	}

	switch strings.ToUpper(pbl.Severity) {
	case "", "ERROR":
		iss.Type = cmdtypes.IssueTypeError
	case "WARNING":
		iss.Type = cmdtypes.IssueTypeWarning
	case "NOTICE":
		iss.Type = cmdtypes.IssueTypeNotice
	default:
		return nil, fmt.Errorf("%s unknown severity %q", skippedIssueMsg, pbl.Severity)
	}

	if !numberRegex.MatchString(pbl.Line) {
		return nil, fmt.Errorf("%s invalid line %q", skippedIssueMsg, pbl.Line)
	}
	iss.Data["line"] = pbl.Line

	if !numberRegex.MatchString(pbl.Column) {
		return nil, fmt.Errorf("%s invalid column %q", skippedIssueMsg, pbl.Column)
	}
	iss.Data["column"] = pbl.Column

	if code := strings.TrimSpace(pbl.Code); code != "" {
		iss.Data["code"] = code
	}

	file := pbl.File
	if file != "" {
		if pbl.FromPath != "" && !filepath.IsAbs(file) {
			file = filepath.Join(pbl.FromPath, file)
		}
		// NOTE: "file" will be refined later by [wire_command.refinePathReporter]
		iss.Data["file"] = file
	}

	return iss, nil
}

func MaskSecret[R any](masker secret.Masker) Middleware[R] {
	return func(sink stream.Sink[R]) stream.Sink[R] {
		return &secretMasker[R]{
			sink:   sink,
			masker: masker,
		}
	}
}

type secretMasker[R any] struct {
	sink   stream.Sink[R]
	masker secret.Masker
}

func (mw *secretMasker[R]) Emit(ctx context.Context, res R, line string) error {
	line = mw.masker.Mask(line)
	return mw.sink.Emit(ctx, res, line)
}
