/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_core

import (
	"fmt"
	"strings"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/wire"
	"drassi.run/gitea-runner/pkg/gitea"
	"go.uber.org/dig"
)

func Module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		if err := scope.Decorate(configureDossier); err != nil {
			return fmt.Errorf("configure records.Dossier: %w", err)
		}
		if err := scope.Decorate(configureSecretMasker); err != nil {
			return fmt.Errorf("configure secret.Masker: %w", err)
		}
		if err := scope.Provide(expressionEnv); err != nil {
		}
		if err := scope.Provide(endpointEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide EnvProvider: %w", err)
		}
		return collectDecorators(scope)
	}
	return wire.NewModule("gitea/core", fn)
}

func configureDossier(task *runnerv1.Task, d *records.Dossier) (*records.Dossier, error) {
	d.Secrets = task.Secrets
	d.Variables = task.Vars
	d.Needs = convertJobNeeds(task.Needs)
	github := d.Github
	if err := model.Decode(task.Context.AsMap(), github); err != nil {
		return nil, err
	} else if github.Token == "" {
		if t := task.Secrets["GITEA_TOKEN"]; t != "" {
			github.Token = t
		} else if t = task.Secrets["GITHUB_TOKEN"]; t != "" {
			github.Token = t
		}
	}
	return d, nil
}

func configureSecretMasker(task *runnerv1.Task, sm secret.Masker) secret.Masker {
	for _, v := range task.Secrets {
		sm.AddSecret(secret.NewValueSecret(v))
	}
	return sm
}

func expressionEnv(d *records.Dossier) (expression.Env, error) {
	opts := []expression.Option{
		expression.WithCache(true),
		expression.WithLibrary(libraries.StdLib()),
		expression.WithVariable("secrets", d.Secrets),
		expression.WithVariable("vars", d.Variables),
		expression.WithVariable("needs", d.Needs),
		expression.WithAlias("gitea", "github"), // make `gitea` variable alias to `github`
		expression.WithVariable("strategy", new(records.Strategy)),
		expression.WithVariable("matrix", make(map[string]string)),
		expression.WithVariable("inputs", make(map[string]any)),
	}
	return expression.NewEnv(opts...)
}

func endpointEnv(client gitea.Client, task *runnerv1.Task, gh *records.Github) executor.EnvProvider {
	endpoint := client.Address()
	endpoint = strings.TrimSuffix(endpoint, "/")

	taskContext := task.Context.Fields
	token := taskContext["gitea_runtime_token"].GetStringValue()
	if token == "" {
		// use task token to action api token for previous Gitea Server Versions
		token = gh.Token
	}

	m := map[string]string{
		"GITEA_ACTIONS":         "true",
		"ACTIONS_RUNTIME_URL":   endpoint + "/api/actions_pipeline/",
		"ACTIONS_RUNTIME_TOKEN": token,
		"ACTIONS_RESULTS_URL":   endpoint,
		//"ACTIONS_CACHE_URL":     "", // TODO
	}
	return executor.StaticEnv(m)
}

func convertJobNeeds(taskNeeds map[string]*runnerv1.TaskNeed) map[string]*records.Need {
	if len(taskNeeds) == 0 {
		return nil
	}

	needs := make(map[string]*records.Need, len(taskNeeds))
	for k, n := range taskNeeds {
		needs[k] = &records.Need{
			Outputs: n.Outputs,
			Result:  resultMap[n.Result],
		}
	}
	return needs
}

var resultMap = map[runnerv1.Result]records.Result{
	runnerv1.Result_RESULT_UNSPECIFIED: "",
	runnerv1.Result_RESULT_SUCCESS:     records.ResultSuccess,
	runnerv1.Result_RESULT_FAILURE:     records.ResultFailure,
	runnerv1.Result_RESULT_CANCELLED:   records.ResultCancelled,
	runnerv1.Result_RESULT_SKIPPED:     records.ResultSkipped,
}
