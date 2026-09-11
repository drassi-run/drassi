/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"net/url"
	"strings"

	"drassi.run/core/pkg/model/records"
)

type ContainerSpec struct {
	Name       string `json:"name,omitempty"`
	Image      string `json:"image,omitempty"`
	PullPolicy string `json:"pull_policy,omitempty"`

	Command     []string          `json:"command,omitempty"`
	Entrypoint  []string          `json:"entrypoint,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	ContainerNetwork  `json:",embed"`
	ContainerStorage  `json:",embed"`
	Devices           []string `json:"devices,omitempty"`
	DeviceCgroupRules []string `json:"device_cgroup_rules,omitempty"`

	ContainerRuntime  `json:",embed"`
	ContainerResource `json:",embed"`
	ContainerSecurity `json:",embed"`
}

const (
	LabelRepository = "run.drassi.repository"
	LabelReference  = "run.drassi.reference"
	LabelWorkflow   = "run.drassi.workflow"
	LabelJob        = "run.drassi.job"
	LabelAttempt    = "run.drassi.attempt"
	LabelRun        = "run.drassi.run"
)

func LabelsFor(forge *records.Forge) map[string]string {
	repo := forge.Repository
	if u, err := url.Parse(forge.ServerUrl); err == nil {
		if server := u.Host; server != "" {
			server = strings.ToLower(server)
			server = strings.TrimRight(server, "/")
			repo = server + "/" + repo
		}
	}

	labels := map[string]string{
		LabelRepository: repo,             // e.g: github.com/drassi-run/drassi
		LabelReference:  forge.Ref,        // e.g: refs/heads/main
		LabelWorkflow:   forge.Workflow,   // e.g: test
		LabelJob:        forge.Job,        // e.g: unittests
		LabelAttempt:    forge.RunAttempt, // e.g: 1
		LabelRun:        forge.RunId,      // e.g: 11208400917
	}
	return labels
}
