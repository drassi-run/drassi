package executor

import "maps"

type EnvProvider interface {
	Env(StepExecutor) map[string]string
}

type StaticEnv map[string]string

func (env StaticEnv) Env(StepExecutor) map[string]string {
	return maps.Clone(env)
}

var CIEnv EnvProvider = StaticEnv{
	"CI":             "true",
	"GITHUB_ACTIONS": "true",
}

type MultiEnvProvider []EnvProvider

func (prov MultiEnvProvider) Env(exec StepExecutor) map[string]string {
	env := make(map[string]string)
	for _, p := range prov {
		maps.Copy(env, p.Env(exec))
	}
	return env
}
