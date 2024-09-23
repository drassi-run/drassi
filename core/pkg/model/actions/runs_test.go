package actions

import (
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	"fmt"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/mapstructure"
	"gotest.tools/v3/assert"
	"reflect"
	"testing"
)

func clone[T any](i T) T {
	if o, err := copystructure.Copy(i); err != nil {
		return i
	} else {
		return o.(T)
	}
}

type runsTestStruct struct {
	Runs          Runs             `actions:"runs"`
	RunsPtr       *Runs            `actions:"runsPtr"`
	ListOfRuns    []Runs           `actions:"listOfRuns"`
	ListOfRunsPtr []*Runs          `actions:"listOfRunsPtr"`
	MapOfRuns     map[string]Runs  `actions:"mapOfRuns"`
	MapOfRunsPtr  map[string]*Runs `actions:"mapOfRunsPtr"`
}

func TestDecodeRuns(t *testing.T) {
	jsInput := map[string]any{
		"using": "node20",
		"main":  "main.js",
	}
	dockerInput := map[string]any{
		"using": "docker",
		"image": "docker://debian:stretch-slim",
	}
	compositeInput := map[string]any{
		"using": "composite",
		"steps": []map[string]any{},
	}

	t.Run("js", func(tt *testing.T) {
		runs := &NodeRuns{
			Using: "node20",
			Main:  "main.js",
		}
		testDecodeRuns(tt, jsInput, runs)
	})
	t.Run("docker", func(tt *testing.T) {
		runs := &DockerRuns{
			Using: "docker",
			Image: "docker://debian:stretch-slim",
		}
		testDecodeRuns(tt, dockerInput, runs)
	})

	t.Run("composite", func(tt *testing.T) {
		runs := &CompositeRuns{
			Using: "composite",
			Steps: []workflows.Step{},
		}
		testDecodeRuns(tt, compositeInput, runs)
	})

	t.Run("conflict/empty", func(tt *testing.T) {
		input := map[string]any{}
		var runs Runs = &NodeRuns{}
		err := model.Decode(input, &runs)

		assert.ErrorContains(tt, err, "`using` is required, and MUST be a string")
	})

	t.Run("conflict/wrong-map-type/1", func(tt *testing.T) {
		var runs Runs = &DockerRuns{}
		err := model.Decode(jsInput, &runs)

		assert.ErrorContains(tt, err, fmt.Sprintf(`map with using=%q CAN'T be decode to %T`, jsInput["using"], runs))
	})

	t.Run("conflict/wrong-map-type/2", func(tt *testing.T) {
		var runs Runs = &CompositeRuns{}
		err := model.Decode(dockerInput, &runs)

		assert.ErrorContains(tt, err, fmt.Sprintf(`map with using=%q CAN'T be decode to %T`, dockerInput["using"], runs))
	})

	t.Run("conflict/wrong-map-type/3", func(tt *testing.T) {
		var runs Runs = &NodeRuns{}
		err := model.Decode(compositeInput, &runs)

		assert.ErrorContains(tt, err, fmt.Sprintf(`map with using=%q CAN'T be decode to %T`, compositeInput["using"], runs))
	})

	t.Run("absent", func(tt *testing.T) {
		type runsStruct struct {
			Runs       Runs            `actions:"runs,omitempty"`
			ListOfRuns []Runs          `actions:"listOfRuns,omitempty"`
			MapOfRuns  map[string]Runs `actions:"mapOfRuns,omitempty"`
		}
		runs := runsStruct{}
		err := model.Decode(map[string]any{}, &runs)

		assert.NilError(tt, err)
		assert.Check(tt, runs.Runs == nil)
		assert.Check(tt, runs.ListOfRuns == nil)
		assert.Check(tt, runs.MapOfRuns == nil)
	})

	t.Run("nil", func(tt *testing.T) {
		type runsStruct struct {
			Runs       Runs            `actions:"runs,omitempty"`
			ListOfRuns []Runs          `actions:"listOfRuns,omitempty"`
			MapOfRuns  map[string]Runs `actions:"mapOfRuns,omitempty"`
		}
		runs := runsStruct{}
		err := model.Decode(map[string]any{
			"runs":       nil,
			"listOfRuns": nil,
			"mapOfRuns":  nil,
		}, &runs)

		assert.NilError(tt, err)
		assert.Check(tt, runs.Runs == nil)
		assert.Check(tt, runs.ListOfRuns == nil)
		assert.Check(tt, runs.MapOfRuns == nil)
	})

	t.Run("invalid-reflection", func(tt *testing.T) {
		opt := func(config *mapstructure.DecoderConfig) {
			fault := func(from reflect.Value, to reflect.Value) (any, error) {
				if !to.Type().Implements(typeRuns) {
					return valueOf(from), nil
				}
				return nil, nil // fault injection
			}
			// DecodeRunsHook MUST be after fault
			config.DecodeHook = mapstructure.ComposeDecodeHookFunc(fault, DecodeRunsHook)
		}

		input := map[string]any{
			"using": "docker",
		}
		runs := new(NodeRuns)
		err := model.DecodeWithOptions(input, &runs, opt)
		assert.NilError(tt, err)
	})
}

func testDecodeRuns[T any](tt *testing.T, value T, runs Runs) {
	data := map[string]any{
		"runs":       clone(value),
		"runsPtr":    clone(value),
		"listOfRuns": []any{clone(value)},
		"mapOfRuns": map[string]any{
			"key": clone(value),
		},
		"listOfRunsPtr": []any{clone(value)},
		"mapOfRunsPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := runsTestStruct{}
	err := model.Decode(data, &actual)

	expected := runsTestStruct{
		Runs:          runs,
		RunsPtr:       &runs,
		ListOfRuns:    []Runs{runs},
		ListOfRunsPtr: []*Runs{&runs},
		MapOfRuns:     map[string]Runs{"key": runs},
		MapOfRunsPtr:  map[string]*Runs{"key": &runs},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected)
}
