package etc

import (
	"drassi.run/core/pkg/executor"
	"go.uber.org/dig"
)

func Wire(scope *dig.Scope) error {
	return scope.Invoke(provideEnv)
}

func provideEnv(sup executor.Supervisor) {
	m := map[string]string{
		"CI":             "true",
		"GITHUB_ACTIONS": "true",
	}
	sup.Register(executor.Env(m))
}
