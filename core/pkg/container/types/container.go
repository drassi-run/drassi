package types

import (
	"net/url"
	"strings"

	"drassi.run/core/pkg/model/records"
)

type ContainerSpec struct {
	Name       string
	Image      string
	PullPolicy string

	Command     []string
	Entrypoint  []string
	WorkingDir  string
	Environment map[string]string
	Labels      map[string]string
	Annotations map[string]string

	ContainerNetwork
	ContainerStorage
	Devices           []string
	DeviceCgroupRules []string

	ContainerRuntime
	ContainerResource
	ContainerSecurity
}

const (
	LabelRepository = "run.drassi.repository"
	LabelReference  = "run.drassi.reference"
	LabelWorkflow   = "run.drassi.workflow"
	LabelJob        = "run.drassi.job"
	LabelAttempt    = "run.drassi.attempt"
	LabelRun        = "run.drassi.run"
)

func LabelsFor(gh *records.Github) map[string]string {
	repo := gh.Repository
	if u, err := url.Parse(gh.ServerUrl); err == nil {
		if server := u.Host; server != "" {
			server = strings.ToLower(server)
			server = strings.TrimRight(server, "/")
			repo = server + "/" + repo
		}
	}

	labels := map[string]string{
		LabelRepository: repo,          // e.g: github.com/drassi-run/drassi
		LabelReference:  gh.Ref,        // e.g: refs/heads/main
		LabelWorkflow:   gh.Workflow,   // e.g: test
		LabelJob:        gh.Job,        // e.g: unittests
		LabelAttempt:    gh.RunAttempt, // e.g: 1
		LabelRun:        gh.RunId,      // e.g: 11208400917
	}
	return labels
}
