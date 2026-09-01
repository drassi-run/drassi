/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_core

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/wire"
	"drassi.run/gitea-runner/pkg/gitea"
	runnerv1 "gitea.dev/actionslib/runner/v1"
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

func convertJobNeeds(taskNeeds map[string]*runnerv1.TaskNeed) map[string]*records.JobResult {
	if len(taskNeeds) == 0 {
		return nil
	}

	needs := make(map[string]*records.JobResult, len(taskNeeds))
	for k, n := range taskNeeds {
		needs[k] = &records.JobResult{
			Result:  resultMap[n.Result],
			Outputs: n.Outputs,
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
