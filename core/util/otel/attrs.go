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
	DrassiWorkflowKey = attribute.Key("drassi.workflow")
	DrassiJobKey      = attribute.Key("drassi.job")
	DrassiStepKey     = attribute.Key("drassi.step")
	DrassiRunKey      = attribute.Key("drassi.run")
	DrassiStageKey    = attribute.Key("drassi.stage")

	ActionPathKey   = attribute.Key("drassi.action.path")
	ActionRepoKey   = attribute.Key("drassi.action.repo")
	ActionScriptKey = attribute.Key("drassi.action.script")

	DrassiCommandKey = attribute.Key("drassi.command")
)

func DrassiWorkflow(s string) attribute.KeyValue {
	return DrassiWorkflowKey.String(s)
}

func DrassiJob(s string) attribute.KeyValue {
	return DrassiJobKey.String(s)
}

func DrassiStep(s string) attribute.KeyValue {
	return DrassiStepKey.String(s)
}

func DrassiRun(s string) attribute.KeyValue {
	return DrassiRunKey.String(s)
}

func DrassiStage[S ~string](s S) attribute.KeyValue {
	return DrassiStageKey.String(string(s))
}

func ActionPath(s string) attribute.KeyValue {
	return ActionPathKey.String(s)
}

func ActionRepo(s string) attribute.KeyValue {
	return ActionRepoKey.String(s)
}

func ActionScript(s string) attribute.KeyValue {
	return ActionScriptKey.String(s)
}

func DrassiCommand(s string) attribute.KeyValue {
	return DrassiCommandKey.String(s)
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
