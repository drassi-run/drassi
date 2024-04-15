package workflows

import (
	"fmt"
	"github.com/dungdm93/drasi/pkg/model"
	"gotest.tools/v3/assert"
	"testing"
)

type stepTestStruct struct {
	Step          Step             `mapstructure:"step"`
	StepPtr       *Step            `mapstructure:"stepPtr"`
	ListOfStep    []Step           `mapstructure:"listOfStep"`
	ListOfStepPtr []*Step          `mapstructure:"listOfStepPtr"`
	MapOfStep     map[string]Step  `mapstructure:"mapOfStep"`
	MapOfStepPtr  map[string]*Step `mapstructure:"mapOfStepPtr"`
}

func TestDecodeStep(t *testing.T) {
	runInput := map[string]any{
		"run": "echo hello world",
	}
	usesInput := map[string]any{
		"uses": "action/pass@v0",
	}

	t.Run("run", func(tt *testing.T) {
		step := &RunStep{
			Run: NewIdent("echo hello world"),
		}
		testDecodeStep(tt, runInput, step)
	})
	t.Run("uses", func(tt *testing.T) {
		step := &UsesStep{
			Uses: "action/pass@v0",
		}
		testDecodeStep(tt, usesInput, step)
	})

	t.Run("conflict/empty", func(tt *testing.T) {
		input := map[string]any{}
		var step Step = &RunStep{}
		err := model.Decode(input, &step)

		assert.ErrorContains(tt, err, "map MUST be contains either `run` or `uses`")
	})

	t.Run("conflict/map-contains-both", func(tt *testing.T) {
		input := map[string]any{
			"run":  "echo hello world",
			"uses": "action/pass@v0",
		}
		var step Step = &RunStep{}
		err := model.Decode(input, &step)

		assert.ErrorContains(tt, err, "map MUST be contains either `run` or `uses`")
	})

	t.Run("conflict/run-to-UsesStep", func(tt *testing.T) {
		var step Step = &UsesStep{}
		err := model.Decode(runInput, &step)

		assert.ErrorContains(tt, err, fmt.Sprintf("map contains `run` CAN'T be decode to %T", step))
	})

	t.Run("conflict/uses-to-RunStep", func(tt *testing.T) {
		var step Step = &RunStep{}
		err := model.Decode(usesInput, &step)

		assert.ErrorContains(tt, err, fmt.Sprintf("map contains `uses` CAN'T be decode to %T", step))
	})

	t.Run("absent", func(tt *testing.T) {
		type stepStruct struct {
			Step       Step            `mapstructure:"step,omitempty"`
			ListOfStep []Step          `mapstructure:"listOfStep,omitempty"`
			MapOfStep  map[string]Step `mapstructure:"mapOfStep,omitempty"`
		}
		step := stepStruct{}
		err := model.Decode(map[string]any{}, &step)

		assert.NilError(tt, err)
		assert.Check(tt, step.Step == nil)
		assert.Check(tt, step.ListOfStep == nil)
		assert.Check(tt, step.MapOfStep == nil)
	})

	t.Run("nil", func(tt *testing.T) {
		type stepStruct struct {
			Step       Step            `mapstructure:"step,omitempty"`
			ListOfStep []Step          `mapstructure:"listOfStep,omitempty"`
			MapOfStep  map[string]Step `mapstructure:"mapOfStep,omitempty"`
		}
		step := stepStruct{}
		err := model.Decode(map[string]any{
			"step":       nil,
			"listOfStep": nil,
			"mapOfStep":  nil,
		}, &step)

		assert.NilError(tt, err)
		assert.Check(tt, step.Step == nil)
		assert.Check(tt, step.ListOfStep == nil)
		assert.Check(tt, step.MapOfStep == nil)
	})
}

func testDecodeStep[T any](tt *testing.T, value T, step Step) {
	data := map[string]any{
		"step":       clone(value),
		"stepPtr":    clone(value),
		"listOfStep": []any{clone(value)},
		"mapOfStep": map[string]any{
			"key": clone(value),
		},
		"listOfStepPtr": []any{clone(value)},
		"mapOfStepPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := stepTestStruct{}
	err := model.Decode(data, &actual)

	opts := commonComparerForEvaluable()
	expected := stepTestStruct{
		Step:          step,
		StepPtr:       &step,
		ListOfStep:    []Step{step},
		ListOfStepPtr: []*Step{&step},
		MapOfStep:     map[string]Step{"key": step},
		MapOfStepPtr:  map[string]*Step{"key": &step},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opts...)
}
