/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_core

import (
	"fmt"
	"regexp"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/common"
	"drassi.run/gha-runner/pkg/messages"
	"go.uber.org/dig"
)

func Module() *wire.Module {
	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(newDossierAndFlags); err != nil {
			return fmt.Errorf("provide records.Dossier and feature.Flags: %w", err)
		}
		if err := scope.Provide(sysEnvProvider, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'sysEnv' records.Dossier: %w", err)
		}
		if err := scope.Decorate(configureSecretMasker); err != nil {
			return fmt.Errorf("configure secret.Masker: %w", err)
		}

		if err := scope.Provide(printSecretSource[executor.JobExecutor], dig.Group(wire.PostStart)); err != nil {
			return fmt.Errorf("provide 'printSecretSource' post-start Hook: %w", err)
		}
		if err := scope.Provide(printTokenPermissions[executor.JobExecutor], dig.Group(wire.PostStart)); err != nil {
			return fmt.Errorf("provide 'printTokenPermissions' post-start Hook: %w", err)
		}

		return collectDecorators(scope)
	}
	return wire.NewModule("gha/core", fn)
}

func configureSecretMasker(req *messages.PipelineAgentJobRequest, sm secret.Masker) (secret.Masker, error) {
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Worker.cs#L140
	for _, v := range req.Variables {
		if v.IsSecret {
			sm.AddSecret(secret.NewValueSecret(v.Value))
		}
	}
	for _, s := range req.MaskHints {
		switch s.Type {
		case messages.MaskTypeVariable:
			sm.AddSecret(secret.NewValueSecret(s.Value))
		case messages.MaskTypeRegex:
			if re, err := regexp.Compile(s.Value); err != nil {
				return nil, fmt.Errorf("invalid regex %q: %w", s.Value, err)
			} else {
				sm.AddSecret(secret.NewRegexSecret(re))
			}
		default:
			return nil, fmt.Errorf("unknown mask type %q", s.Type)
		}
	}
	if res := req.Resources; res != nil {
		for _, ep := range res.Endpoints {
			authz := ep.Authorization
			if authz == nil {
				continue
			}
			for _, v := range authz.Parameters {
				if v != "" {
					sm.AddSecret(secret.NewValueSecret(v))
				}
			}
		}
	}

	return sm, nil
}

func sysEnvProvider(req *messages.PipelineAgentJobRequest) (executor.EnvProvider, error) {
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Handlers/NodeScriptActionHandler.cs#L53-L78
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L218-L238
	sysCon := req.ServiceEndpoint("SystemVssConnection")
	if sysCon == nil {
		return nil, fmt.Errorf("service endpoint 'SystemVssConnection' not found")
	}
	var accessToken string
	if authz := sysCon.Authorization; authz != nil && authz.Scheme == "OAuth" {
		accessToken = authz.Parameters["AccessToken"]
	}
	sysEnv := map[string]string{
		"ACTIONS_RUNTIME_URL":   sysCon.Url,
		"ACTIONS_RUNTIME_TOKEN": accessToken,
	}
	if url := sysCon.Data["CacheServerUrl"]; url != "" {
		sysEnv["ACTIONS_CACHE_URL"] = url
	}
	if cacheV2 := req.Variables["actions_uses_cache_service_v2"]; strings.ToLower(cacheV2.Value) == "true" {
		sysEnv["ACTIONS_CACHE_SERVICE_V2"] = "True" // bool.TrueString
	}
	if url := sysCon.Data["PipelinesServiceUrl"]; url != "" {
		sysEnv["ACTIONS_RUNTIME_URL"] = url
	}
	if url := sysCon.Data["GenerateIdTokenUrl"]; url != "" {
		sysEnv["ACTIONS_ID_TOKEN_REQUEST_URL"] = url
		sysEnv["ACTIONS_ID_TOKEN_REQUEST_TOKEN"] = accessToken
	}
	if url := sysCon.Data["ResultsServiceUrl"]; url != "" {
		sysEnv["ACTIONS_RESULTS_URL"] = url
	} else if v, ok := req.Variables["system.github.results_endpoint"]; ok {
		sysEnv["ACTIONS_RESULTS_URL"] = v.Value
	}

	return executor.StaticEnv(sysEnv), nil
}

func newDossierAndFlags(req *messages.PipelineAgentJobRequest, runner *records.RunnerInfo) (*records.Dossier, feature.Flags, error) {
	dossier := new(records.Dossier)
	if err := model.Decode(req.ContextData, dossier); err != nil {
		return nil, nil, fmt.Errorf("decode ContextData: %w", err)
	}

	if dossier.Forge == nil {
		dossier.Forge = new(records.Forge)
	}
	if dossier.Env == nil {
		dossier.Env = make(map[string]string)
	}
	dossier.Runner = runner

	// GitHub context
	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/ExecutionContext.cs#L882-L891
	forge := dossier.Forge
	if forge.Token == "" {
		if v, ok := req.Variables["system.github.token"]; ok {
			forge.Token = v.Value
		} else if v, ok = req.Variables["github_token"]; ok {
			forge.Token = v.Value
		}
	}
	if forge.Job == "" {
		if v, ok := req.Variables["system.github.job"]; ok {
			forge.Job = v.Value
		}
	}

	// Sets debug using vars context in case debug variables are not present.
	// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/ExecutionContext.cs#L1394-L1417
	sysVar := req.Variables
	dVar := dossier.Variables
	if _, ok := sysVar[wire.RunnerDebug]; !ok {
		if v, ok := dVar[wire.RunnerDebug]; ok {
			sysVar[wire.RunnerDebug] = messages.Variable{Value: v}
		}
	}
	if _, ok := sysVar[wire.StepDebug]; !ok {
		if v, ok := dVar[wire.StepDebug]; ok {
			sysVar[wire.StepDebug] = messages.Variable{Value: v}
		}
	}
	flags := common.NewSysVarFlags(req.Variables)

	if feature.Bool(flags, wire.RunnerDebug, false) {
		runner.Debug = "1"
	}

	return dossier, flags, nil
}
