package workflows

import (
	"drassi.run/core/pkg/model"
	utilreflect "drassi.run/core/pkg/util/reflect"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"gotest.tools/v3/assert"
	"reflect"
	"testing"
)

type jobTestStruct struct {
	Job          Job             `actions:"job"`
	JobPtr       *Job            `actions:"jobPtr"`
	ListOfJob    []Job           `actions:"listOfJob"`
	ListOfJobPtr []*Job          `actions:"listOfJobPtr"`
	MapOfJob     map[string]Job  `actions:"mapOfJob"`
	MapOfJobPtr  map[string]*Job `actions:"mapOfJobPtr"`
}

func TestDecodeJob(t *testing.T) {
	runsOnInput := map[string]any{
		"runs-on": "ubuntu",
	}
	usesInput := map[string]any{
		"uses": "./path/to/the/workflow.yaml",
	}

	t.Run("normal", func(tt *testing.T) {
		job := &NormalJob{
			RunsOn: Evaluable[RunsOn]{
				Token: NewLiteralToken("ubuntu"),
			},
		}
		testDecodeJob(tt, runsOnInput, job)
	})
	t.Run("reusableWorkflowCall", func(tt *testing.T) {
		job := &ReusableWorkflowCallJob{
			Uses: "./path/to/the/workflow.yaml",
		}
		testDecodeJob(tt, usesInput, job)
	})

	t.Run("conflict/empty", func(tt *testing.T) {
		input := map[string]any{}
		var job Job = &NormalJob{}
		err := model.Decode(input, &job)

		assert.ErrorContains(tt, err, "map MUST be contains either `runs-on` or `uses`")
	})

	t.Run("conflict/map-contains-both", func(tt *testing.T) {
		input := map[string]any{
			"runs-on": "ubuntu",
			"uses":    "./path/to/the/workflow.yaml",
		}
		var job Job = &NormalJob{}
		err := model.Decode(input, &job)

		assert.ErrorContains(tt, err, "map MUST be contains either `runs-on` or `uses`")
	})

	t.Run("conflict/runs-on-ReusableWorkflowCallJob", func(tt *testing.T) {
		var job Job = &ReusableWorkflowCallJob{}
		err := model.Decode(runsOnInput, &job)

		assert.ErrorContains(tt, err, fmt.Sprintf("map contains `runs-on` CAN'T be decode to %T", job))
	})

	t.Run("conflict/uses-to-NormalJob", func(tt *testing.T) {
		var job Job = &NormalJob{}
		err := model.Decode(usesInput, &job)

		assert.ErrorContains(tt, err, fmt.Sprintf("map contains `uses` CAN'T be decode to %T", job))
	})

	t.Run("absent", func(tt *testing.T) {
		type jobStruct struct {
			Job       Job            `actions:"job,omitempty"`
			ListOfJob []Job          `actions:"listOfJob,omitempty"`
			MapOfJob  map[string]Job `actions:"mapOfJob,omitempty"`
		}
		job := jobStruct{}
		err := model.Decode(map[string]any{}, &job)

		assert.NilError(tt, err)
		assert.Check(tt, job.Job == nil)
		assert.Check(tt, job.ListOfJob == nil)
		assert.Check(tt, job.MapOfJob == nil)
	})

	t.Run("nil", func(tt *testing.T) {
		type jobStruct struct {
			Job       Job            `actions:"job,omitempty"`
			ListOfJob []Job          `actions:"listOfJob,omitempty"`
			MapOfJob  map[string]Job `actions:"mapOfJob,omitempty"`
		}
		job := jobStruct{}
		err := model.Decode(map[string]any{
			"job":       nil,
			"listOfJob": nil,
			"mapOfJob":  nil,
		}, &job)

		assert.NilError(tt, err)
		assert.Check(tt, job.Job == nil)
		assert.Check(tt, job.ListOfJob == nil)
		assert.Check(tt, job.MapOfJob == nil)
	})

	t.Run("invalid-reflection", func(tt *testing.T) {
		opt := func(config *mapstructure.DecoderConfig) {
			fault := func(from reflect.Value, to reflect.Value) (any, error) {
				if !to.Type().Implements(typeJob) {
					return utilreflect.ValueOf(from), nil
				}
				return nil, nil // fault injection
			}
			// DecodeJobHook MUST be after fault
			config.DecodeHook = mapstructure.ComposeDecodeHookFunc(fault, DecodeJobHook)
		}

		input := map[string]any{
			"runs-on": "ubuntu",
		}
		job := new(NormalJob)
		err := model.DecodeWithOptions(input, &job, opt)
		assert.NilError(tt, err)
	})
}

func testDecodeJob[T any](tt *testing.T, value T, job Job) {
	data := map[string]any{
		"job":       clone(value),
		"jobPtr":    clone(value),
		"listOfJob": []any{clone(value)},
		"mapOfJob": map[string]any{
			"key": clone(value),
		},
		"listOfJobPtr": []any{clone(value)},
		"mapOfJobPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := jobTestStruct{}
	err := model.Decode(data, &actual)

	opts := comparerForLiteralToken()
	expected := jobTestStruct{
		Job:          job,
		JobPtr:       &job,
		ListOfJob:    []Job{job},
		ListOfJobPtr: []*Job{&job},
		MapOfJob:     map[string]Job{"key": job},
		MapOfJobPtr:  map[string]*Job{"key": &job},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opts)
}

type jobNeedsTestStruct struct {
	JN          JobNeeds             `actions:"jn"`
	JNPtr       *JobNeeds            `actions:"jnPtr"`
	ListOfJN    []JobNeeds           `actions:"listOfJn"`
	MapOfJN     map[string]JobNeeds  `actions:"mapOfJn"`
	ListOfJNPtr []*JobNeeds          `actions:"listOfJnPtr"`
	MapOfJNPtr  map[string]*JobNeeds `actions:"mapOfJnPtr"`
}

func TestDecodeJobNeeds(t *testing.T) {
	t.Run("single", func(tt *testing.T) {
		jn := JobNeeds{"single"}
		testDecodeJobNeeds(tt, "single", jn)
	})
	t.Run("multi", func(tt *testing.T) {
		jn := JobNeeds{"first", "second", "third"}
		testDecodeJobNeeds(tt, []string{"first", "second", "third"}, jn)
	})
}

func testDecodeJobNeeds[T any](tt *testing.T, value T, jn JobNeeds) {
	data := map[string]any{
		"jn":       clone(value),
		"jnPtr":    clone(value),
		"listOfJn": []any{clone(value)},
		"mapOfJn": map[string]any{
			"key": clone(value),
		},
		"listOfJnPtr": []any{clone(value)},
		"mapOfJnPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := jobNeedsTestStruct{}
	err := model.Decode(data, &actual)

	expected := jobNeedsTestStruct{
		JN:          jn,
		JNPtr:       &jn,
		ListOfJN:    []JobNeeds{jn},
		MapOfJN:     map[string]JobNeeds{"key": jn},
		ListOfJNPtr: []*JobNeeds{&jn},
		MapOfJNPtr:  map[string]*JobNeeds{"key": &jn},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected)
}

type jobSecretTestStruct struct {
	JS          JobSecrets             `actions:"js"`
	JSPtr       *JobSecrets            `actions:"jsPtr"`
	ListOfJS    []JobSecrets           `actions:"listOfJs"`
	MapOfJS     map[string]JobSecrets  `actions:"mapOfJs"`
	ListOfJSPtr []*JobSecrets          `actions:"listOfJsPtr"`
	MapOfJSPtr  map[string]*JobSecrets `actions:"mapOfJsPtr"`
}

func TestDecodeJobSecrets(t *testing.T) {
	t.Run("inherit", func(tt *testing.T) {
		js := JobSecrets{
			Inherit: true,
		}
		testDecodeJobSecrets(tt, "inherit", js)
	})

	t.Run("map/string", func(tt *testing.T) {
		js := JobSecrets{
			Secrets: map[string]string{
				"first":  "foobar",
				"second": "abcyxz",
			},
		}
		v := map[string]string{
			"first":  "foobar",
			"second": "abcyxz",
		}
		testDecodeJobSecrets(tt, v, js)
	})
	t.Run("map/any", func(tt *testing.T) {
		js := JobSecrets{
			Secrets: map[string]string{
				"first":  "foobar",
				"second": "abcyxz",
			},
		}
		v := map[string]any{
			"first":  "foobar",
			"second": "abcyxz",
		}
		testDecodeJobSecrets(tt, v, js)
	})
}

func testDecodeJobSecrets[T any](tt *testing.T, value T, js JobSecrets) {
	data := map[string]any{
		"js":       clone(value),
		"jsPtr":    clone(value),
		"listOfJs": []any{clone(value)},
		"mapOfJs": map[string]any{
			"key": clone(value),
		},
		"listOfJsPtr": []any{clone(value)},
		"mapOfJsPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := jobSecretTestStruct{}
	err := model.Decode(data, &actual)

	expected := jobSecretTestStruct{
		JS:          js,
		JSPtr:       &js,
		ListOfJS:    []JobSecrets{js},
		MapOfJS:     map[string]JobSecrets{"key": js},
		ListOfJSPtr: []*JobSecrets{&js},
		MapOfJSPtr:  map[string]*JobSecrets{"key": &js},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected)
}

type environmentTestStruct struct {
	Environment          Environment             `actions:"environment"`
	EnvironmentPtr       *Environment            `actions:"environmentPtr"`
	ListOfEnvironment    []Environment           `actions:"listOfEnvironment"`
	MapOfEnvironment     map[string]Environment  `actions:"mapOfEnvironment"`
	ListOfEnvironmentPtr []*Environment          `actions:"listOfEnvironmentPtr"`
	MapOfEnvironmentPtr  map[string]*Environment `actions:"mapOfEnvironmentPtr"`
}

func TestDecodeEnvironment(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		env := Environment{
			Name: "name1",
		}
		testDecodeEnvironment(tt, "name1", env)
	})

	t.Run("map", func(tt *testing.T) {
		env := Environment{
			Name: "name1",
			Url:  "https://example.com",
		}
		v := map[string]any{
			"name": "name1",
			"url":  "https://example.com",
		}
		testDecodeEnvironment(tt, v, env)
	})
}

func testDecodeEnvironment[T any](tt *testing.T, value T, env Environment) {
	data := map[string]any{
		"environment":       clone(value),
		"environmentPtr":    clone(value),
		"listOfEnvironment": []any{clone(value)},
		"mapOfEnvironment": map[string]any{
			"key": clone(value),
		},
		"listOfEnvironmentPtr": []any{clone(value)},
		"mapOfEnvironmentPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := environmentTestStruct{}
	err := model.Decode(data, &actual)

	expected := environmentTestStruct{
		Environment:          env,
		EnvironmentPtr:       &env,
		ListOfEnvironment:    []Environment{env},
		MapOfEnvironment:     map[string]Environment{"key": env},
		ListOfEnvironmentPtr: []*Environment{&env},
		MapOfEnvironmentPtr:  map[string]*Environment{"key": &env},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected)
}

type runOnTestStruct struct {
	RO          RunsOn             `actions:"ro"`
	ROPtr       *RunsOn            `actions:"roPtr"`
	ListOfRO    []RunsOn           `actions:"listOfRo"`
	MapOfRO     map[string]RunsOn  `actions:"mapOfRo"`
	ListOfROPtr []*RunsOn          `actions:"listOfRoPtr"`
	MapOfROPtr  map[string]*RunsOn `actions:"mapOfRoPtr"`
}

func TestDecodeRunsOn(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []string{"label1"},
		}
		testDecodeRunsOn(tt, "label1", ro)
	})

	t.Run("list/string", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []string{"label1", "label2"},
		}
		testDecodeRunsOn(tt, []string{"label1", "label2"}, ro)
	})

	t.Run("list/any", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []string{"label1", "label2"},
		}
		testDecodeRunsOn(tt, []any{"label1", "label2"}, ro)
	})

	t.Run("map/single_label", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []string{"label1"},
		}
		v := map[string]any{
			"labels": "label1",
		}
		testDecodeRunsOn(tt, v, ro)
	})

	t.Run("map/list_labels", func(tt *testing.T) {
		ro := RunsOn{
			Group:  "foobar",
			Labels: []string{"label1", "label1"},
		}
		v := map[string]any{
			"group":  "foobar",
			"labels": []string{"label1", "label1"},
		}
		testDecodeRunsOn(tt, v, ro)
	})
}

func testDecodeRunsOn[T any](tt *testing.T, value T, ro RunsOn) {
	data := map[string]any{
		"ro":       clone(value),
		"roPtr":    clone(value),
		"listOfRo": []any{clone(value)},
		"mapOfRo": map[string]any{
			"key": clone(value),
		},
		"listOfRoPtr": []any{clone(value)},
		"mapOfRoPtr": map[string]any{
			"key": clone(value),
		},
	}

	actual := runOnTestStruct{}
	err := model.Decode(data, &actual)

	expected := runOnTestStruct{
		RO:          ro,
		ROPtr:       &ro,
		ListOfRO:    []RunsOn{ro},
		MapOfRO:     map[string]RunsOn{"key": ro},
		ListOfROPtr: []*RunsOn{&ro},
		MapOfROPtr:  map[string]*RunsOn{"key": &ro},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected)
}
