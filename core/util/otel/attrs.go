/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xotel

import (
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
)

var (
	RepoKey     = attribute.Key("drassi.repo")
	WorkflowKey = attribute.Key("drassi.workflow")
	JobKey      = attribute.Key("drassi.job")
	StepKey     = attribute.Key("drassi.step")
	RunKey      = attribute.Key("drassi.run")

	ActionRepoKey   = attribute.Key("drassi.action.repo")
	ActionPathKey   = attribute.Key("drassi.action.path")
	ActionScriptKey = attribute.Key("drassi.action.script")

	CommandKey = attribute.Key("drassi.command")
)

func Repo(s string) attribute.KeyValue {
	return RepoKey.String(s)
}

func Workflow(s string) attribute.KeyValue {
	return WorkflowKey.String(s)
}

func Job(s string) attribute.KeyValue {
	return JobKey.String(s)
}

func Step(s string) attribute.KeyValue {
	return StepKey.String(s)
}

func Run(s string) attribute.KeyValue {
	return RunKey.String(s)
}

func ActionRepo(s string) attribute.KeyValue {
	return ActionRepoKey.String(s)
}

func ActionPath(s string) attribute.KeyValue {
	return ActionPathKey.String(s)
}

func ActionScript(s string) attribute.KeyValue {
	return ActionScriptKey.String(s)
}

func Command(s string) attribute.KeyValue {
	return CommandKey.String(s)
}

func ToSlogAttrs(kv ...attribute.KeyValue) []slog.Attr {
	attrs := make([]slog.Attr, 0)
	for _, attr := range kv {
		k, v := string(attr.Key), attr.Value
		switch attr.Value.Type() {
		case attribute.BOOL:
			attrs = append(attrs, slog.Bool(k, v.AsBool()))
		case attribute.INT64:
			attrs = append(attrs, slog.Int64(k, v.AsInt64()))
		case attribute.FLOAT64:
			attrs = append(attrs, slog.Float64(k, v.AsFloat64()))
		case attribute.STRING:
			attrs = append(attrs, slog.String(k, v.AsString()))
		default:
			// ignore
		}
	}
	return attrs
}
