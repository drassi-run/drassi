/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"fmt"
	"reflect"
	"testing"

	"drassi.run/core/pkg/model"
	"github.com/go-viper/mapstructure/v2"
	"gotest.tools/v3/assert"
)

type stepTestStruct struct {
	Step          Step             `actions:"step"`
	StepPtr       *Step            `actions:"stepPtr"`
	ListOfStep    []Step           `actions:"listOfStep"`
	ListOfStepPtr []*Step          `actions:"listOfStepPtr"`
	MapOfStep     map[string]Step  `actions:"mapOfStep"`
	MapOfStepPtr  map[string]*Step `actions:"mapOfStepPtr"`
}

func TestDecodeStep(t *testing.T) {
	runInput := map[string]any{
		"run": "echo hello world",
	}
	usesInput := map[string]any{
		"uses": "action/pass@v0",
	}

	t.Run("run", func(tt *testing.T) {
		step := &RunActionStep{
			Run: NewLiteralToken("echo hello world"),
		}
		testDecodeStep(tt, runInput, step)
	})
	t.Run("uses", func(tt *testing.T) {
		step := &UsesActionStep{
			Uses: "action/pass@v0",
		}
		testDecodeStep(tt, usesInput, step)
	})

	t.Run("conflict/empty", func(tt *testing.T) {
		input := map[string]any{}
		var step Step = &RunActionStep{}
		err := model.Decode(input, &step)

		assert.ErrorContains(tt, err, "map MUST be contains either `run` or `uses`")
	})

	t.Run("conflict/map-contains-both", func(tt *testing.T) {
		input := map[string]any{
			"run":  "echo hello world",
			"uses": "action/pass@v0",
		}
		var step Step = &RunActionStep{}
		err := model.Decode(input, &step)

		assert.ErrorContains(tt, err, "map MUST be contains either `run` or `uses`")
	})

	t.Run("conflict/run-to-UsesStep", func(tt *testing.T) {
		var step Step = &UsesActionStep{}
		err := model.Decode(runInput, &step)

		assert.ErrorContains(tt, err, fmt.Sprintf("map contains `run` CAN'T be decode to %T", step))
	})

	t.Run("conflict/uses-to-RunStep", func(tt *testing.T) {
		var step Step = &RunActionStep{}
		err := model.Decode(usesInput, &step)

		assert.ErrorContains(tt, err, fmt.Sprintf("map contains `uses` CAN'T be decode to %T", step))
	})

	t.Run("absent", func(tt *testing.T) {
		type stepStruct struct {
			Step       Step            `actions:"step,omitempty"`
			ListOfStep []Step          `actions:"listOfStep,omitempty"`
			MapOfStep  map[string]Step `actions:"mapOfStep,omitempty"`
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
			Step       Step            `actions:"step,omitempty"`
			ListOfStep []Step          `actions:"listOfStep,omitempty"`
			MapOfStep  map[string]Step `actions:"mapOfStep,omitempty"`
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

	t.Run("invalid-reflection", func(tt *testing.T) {
		opt := func(config *mapstructure.DecoderConfig) {
			fault := func(from reflect.Value, to reflect.Value) (any, error) {
				if !to.Type().Implements(typeStep) {
					return valueOf(from), nil
				}
				return nil, nil // fault injection
			}
			// DecodeStepHook MUST be after fault
			config.DecodeHook = mapstructure.ComposeDecodeHookFunc(fault, DecodeStepHook)
		}

		input := map[string]any{
			"run": "true",
		}
		step := new(RunActionStep)
		err := model.DecodeWithOptions(input, &step, opt)
		assert.NilError(tt, err)
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

	opts := comparerForLiteralToken()
	expected := stepTestStruct{
		Step:          step,
		StepPtr:       &step,
		ListOfStep:    []Step{step},
		ListOfStepPtr: []*Step{&step},
		MapOfStep:     map[string]Step{"key": step},
		MapOfStepPtr:  map[string]*Step{"key": &step},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opts)
}
