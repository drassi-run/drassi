package contexts

import (
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/model"
)

// The `runner` context contains information about the runner that is executing the current job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
type Runner struct {
	Name      string             `json:"name" yaml:"name"`
	Os        model.Machine      `json:"os" yaml:"os"`
	Arch      model.Architecture `json:"arch" yaml:"arch"`
	Temp      string             `json:"temp" yaml:"temp"`
	ToolCache string             `json:"tool_cache" yaml:"tool_cache"`
	Debug     string             `json:"debug" yaml:"debug"`
}

var RunnerFields map[string]int

func init() {
	t := reflect.TypeFor[Runner]()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		v := strings.Split(f.Tag.Get("json"), ",")[0]
		if v == "" || v == "-" {
			continue
		}
		RunnerFields[v] = i
	}
}
