package etc

import (
	"drassi.run/core/pkg/executor"
	"go.uber.org/dig"
)

func Wire(scope *dig.Scope) error {
	return scope.Invoke(provideEnv)
}

func provideEnv(sup executor.Supervisor) {
	fn := func() map[string]string {
		m := map[string]string{
			"CI":             "true",
			"GITHUB_ACTIONS": "true",
		}
		return m
	}
	sup.Register(executor.EnvProvider(fn))
}
