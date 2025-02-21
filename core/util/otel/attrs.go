package xotel

import "go.opentelemetry.io/otel/attribute"

var (
	DrassiWorkflowKey = attribute.Key("drassi.workflow")
	DrassiJobKey      = attribute.Key("drassi.job")
	DrassiStepKey     = attribute.Key("drassi.step")
	DrassiRunKey      = attribute.Key("drassi.run")
	DrassiStageKey    = attribute.Key("drassi.stage")

	ActionPathKey   = attribute.Key("drassi.action.path")
	ActionRepoKey   = attribute.Key("drassi.action.repo")
	ActionScriptKey = attribute.Key("drassi.action.script")
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

func DrassiStage(s string) attribute.KeyValue {
	return DrassiStageKey.String(s)
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
