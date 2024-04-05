package workflows

import (
	"fmt"
	"github.com/dungdm93/drasi/pkg/model"
	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	"gotest.tools/v3/assert"
	"testing"
)

func clone[T any](i any) any {
	if o, err := copystructure.Copy(i); err != nil {
		return i
	} else {
		return o.(T)
	}
}

type jobTestStruct struct {
	Job          Job             `mapstructure:"job"`
	JobPtr       *Job            `mapstructure:"jobPtr"`
	ListOfJob    []Job           `mapstructure:"listOfJob"`
	ListOfJobPtr []*Job          `mapstructure:"listOfJobPtr"`
	MapOfJob     map[string]Job  `mapstructure:"mapOfJob"`
	MapOfJobPtr  map[string]*Job `mapstructure:"mapOfJobPtr"`
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
			RunsOn: RunsOn{
				Labels: []Evaluable[string]{newIdent("ubuntu")},
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

	t.Run("nil", func(tt *testing.T) {
		type jobStruct struct {
			Job       Job            `mapstructure:"job,omitempty"`
			ListOfJob []Job          `mapstructure:"listOfJob,omitempty"`
			MapOfJob  map[string]Job `mapstructure:"mapOfJob,omitempty"`
		}
		job := jobStruct{}
		err := model.Decode(map[string]any{}, &job)

		assert.NilError(tt, err)
		assert.Check(tt, job.Job == nil)
		assert.Check(tt, job.ListOfJob == nil)
		assert.Check(tt, job.MapOfJob == nil)
	})
}

func testDecodeJob[T any](tt *testing.T, value T, job Job) {
	data := map[string]any{
		"job":       clone[T](value),
		"jobPtr":    clone[T](value),
		"listOfJob": []any{clone[T](value)},
		"mapOfJob": map[string]any{
			"key": clone[T](value),
		},
		"listOfJobPtr": []any{clone[T](value)},
		"mapOfJobPtr": map[string]any{
			"key": clone[T](value),
		},
	}

	actual := jobTestStruct{}
	err := model.Decode(data, &actual)

	opt := []cmp.Option{
		comparerForEvaluable[string](),
		comparerForEvaluable[bool](),
		comparerForEvaluable[int64](),
		comparerForEvaluable[float64](),
	}
	expected := jobTestStruct{
		Job:          job,
		JobPtr:       &job,
		ListOfJob:    []Job{job},
		ListOfJobPtr: []*Job{&job},
		MapOfJob:     map[string]Job{"key": job},
		MapOfJobPtr:  map[string]*Job{"key": &job},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opt...)
}

type jobNeedsTestStruct struct {
	JN          JobNeeds             `mapstructure:"jn"`
	JNPtr       *JobNeeds            `mapstructure:"jnPtr"`
	ListOfJN    []JobNeeds           `mapstructure:"listOfJn"`
	MapOfJN     map[string]JobNeeds  `mapstructure:"mapOfJn"`
	ListOfJNPtr []*JobNeeds          `mapstructure:"listOfJnPtr"`
	MapOfJNPtr  map[string]*JobNeeds `mapstructure:"mapOfJnPtr"`
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
		"jn":       clone[T](value),
		"jnPtr":    clone[T](value),
		"listOfJn": []any{clone[T](value)},
		"mapOfJn": map[string]any{
			"key": clone[T](value),
		},
		"listOfJnPtr": []any{clone[T](value)},
		"mapOfJnPtr": map[string]any{
			"key": clone[T](value),
		},
	}

	actual := jobNeedsTestStruct{}
	err := model.Decode(data, &actual)

	opt := comparerForEvaluable[string]()
	expected := jobNeedsTestStruct{
		JN:          jn,
		JNPtr:       &jn,
		ListOfJN:    []JobNeeds{jn},
		MapOfJN:     map[string]JobNeeds{"key": jn},
		ListOfJNPtr: []*JobNeeds{&jn},
		MapOfJNPtr:  map[string]*JobNeeds{"key": &jn},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opt)
}

type jobSecretTestStruct struct {
	JS          JobSecrets             `mapstructure:"js"`
	JSPtr       *JobSecrets            `mapstructure:"jsPtr"`
	ListOfJS    []JobSecrets           `mapstructure:"listOfJs"`
	MapOfJS     map[string]JobSecrets  `mapstructure:"mapOfJs"`
	ListOfJSPtr []*JobSecrets          `mapstructure:"listOfJsPtr"`
	MapOfJSPtr  map[string]*JobSecrets `mapstructure:"mapOfJsPtr"`
}

func TestDecodeJobSecrets(t *testing.T) {
	t.Run("inherit", func(tt *testing.T) {
		js := JobSecrets{
			Inherit: true,
		}
		testDecodeJobSecrets(tt, "inherit", js)
	})

	t.Run("expr", func(tt *testing.T) {
		js := JobSecrets{
			Secrets: map[string]Evaluable[string]{
				"normal":     newIdent("normal"),
				"expression": newExpr("${{ expression }}", toString),
			},
		}
		v := map[string]string{
			"normal":     "normal",
			"expression": "${{ expression }}",
		}
		testDecodeJobSecrets(tt, v, js)
	})
}

func testDecodeJobSecrets[T any](tt *testing.T, value T, js JobSecrets) {
	data := map[string]any{
		"js":       clone[T](value),
		"jsPtr":    clone[T](value),
		"listOfJs": []any{clone[T](value)},
		"mapOfJs": map[string]any{
			"key": clone[T](value),
		},
		"listOfJsPtr": []any{clone[T](value)},
		"mapOfJsPtr": map[string]any{
			"key": clone[T](value),
		},
	}

	actual := jobSecretTestStruct{}
	err := model.Decode(data, &actual)

	opt := comparerForEvaluable[string]()
	expected := jobSecretTestStruct{
		JS:          js,
		JSPtr:       &js,
		ListOfJS:    []JobSecrets{js},
		MapOfJS:     map[string]JobSecrets{"key": js},
		ListOfJSPtr: []*JobSecrets{&js},
		MapOfJSPtr:  map[string]*JobSecrets{"key": &js},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opt)
}

type environmentTestStruct struct {
	Environment          Environment             `mapstructure:"environment"`
	EnvironmentPtr       *Environment            `mapstructure:"environmentPtr"`
	ListOfEnvironment    []Environment           `mapstructure:"listOfEnvironment"`
	MapOfEnvironment     map[string]Environment  `mapstructure:"mapOfEnvironment"`
	ListOfEnvironmentPtr []*Environment          `mapstructure:"listOfEnvironmentPtr"`
	MapOfEnvironmentPtr  map[string]*Environment `mapstructure:"mapOfEnvironmentPtr"`
}

func TestDecodeEnvironment(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		env := Environment{
			Name: newIdent("name1"),
		}
		testDecodeEnvironment(tt, "name1", env)
	})

	t.Run("expr", func(tt *testing.T) {
		env := Environment{
			Name: newExpr("${{ foobar }}", toString),
		}
		testDecodeEnvironment(tt, "${{ foobar }}", env)
	})

	t.Run("map/string", func(tt *testing.T) {
		env := Environment{
			Name: newIdent("name1"),
			Url:  newExpr("${{ foobar }}", toString),
		}
		v := map[string]any{
			"name": "name1",
			"url":  "${{ foobar }}",
		}
		testDecodeEnvironment(tt, v, env)
	})

	t.Run("map/expr", func(tt *testing.T) {
		env := Environment{
			Name: newExpr("${{ foobar }}", toString),
			Url:  newIdent("url1"),
		}
		v := map[string]any{
			"name": "${{ foobar }}",
			"url":  "url1",
		}
		testDecodeEnvironment(tt, v, env)
	})
}

func testDecodeEnvironment[T any](tt *testing.T, value T, env Environment) {
	data := map[string]any{
		"environment":       clone[T](value),
		"environmentPtr":    clone[T](value),
		"listOfEnvironment": []any{clone[T](value)},
		"mapOfEnvironment": map[string]any{
			"key": clone[T](value),
		},
		"listOfEnvironmentPtr": []any{clone[T](value)},
		"mapOfEnvironmentPtr": map[string]any{
			"key": clone[T](value),
		},
	}

	actual := environmentTestStruct{}
	err := model.Decode(data, &actual)

	opt := comparerForEvaluable[string]()
	expected := environmentTestStruct{
		Environment:          env,
		EnvironmentPtr:       &env,
		ListOfEnvironment:    []Environment{env},
		MapOfEnvironment:     map[string]Environment{"key": env},
		ListOfEnvironmentPtr: []*Environment{&env},
		MapOfEnvironmentPtr:  map[string]*Environment{"key": &env},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opt)
}

type runOnTestStruct struct {
	RO          RunsOn             `mapstructure:"ro"`
	ROPtr       *RunsOn            `mapstructure:"roPtr"`
	ListOfRO    []RunsOn           `mapstructure:"listOfRo"`
	MapOfRO     map[string]RunsOn  `mapstructure:"mapOfRo"`
	ListOfROPtr []*RunsOn          `mapstructure:"listOfRoPtr"`
	MapOfROPtr  map[string]*RunsOn `mapstructure:"mapOfRoPtr"`
}

func TestDecodeRunsOn(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []Evaluable[string]{newIdent("label1")},
		}
		testDecodeRunsOn(tt, "label1", ro)
	})

	t.Run("expr", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []Evaluable[string]{newExpr("${{ foobar }}", toString)},
		}
		testDecodeRunsOn(tt, "${{ foobar }}", ro)
	})

	t.Run("list", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []Evaluable[string]{
				newIdent("label1"),
				newExpr("${{ foobar }}", toString),
			},
		}
		testDecodeRunsOn(tt, []string{"label1", "${{ foobar }}"}, ro)
	})

	t.Run("map/string", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []Evaluable[string]{newIdent("label1")},
		}
		v := map[string]any{
			"labels": "label1",
		}
		testDecodeRunsOn(tt, v, ro)
	})

	t.Run("map/expr", func(tt *testing.T) {
		ro := RunsOn{
			Labels: []Evaluable[string]{newExpr("${{ foobar }}", toString)},
		}
		v := map[string]any{
			"labels": "${{ foobar }}",
		}
		testDecodeRunsOn(tt, v, ro)
	})

	t.Run("map/list", func(tt *testing.T) {
		ro := RunsOn{
			Group: "foobar",
			Labels: []Evaluable[string]{
				newIdent("label1"),
				newExpr("${{ foobar }}", toString),
			},
		}
		v := map[string]any{
			"group":  "foobar",
			"labels": []string{"label1", "${{ foobar }}"},
		}
		testDecodeRunsOn(tt, v, ro)
	})
}

func testDecodeRunsOn[T any](tt *testing.T, value T, ro RunsOn) {
	data := map[string]any{
		"ro":       clone[T](value),
		"roPtr":    clone[T](value),
		"listOfRo": []any{clone[T](value)},
		"mapOfRo": map[string]any{
			"key": clone[T](value),
		},
		"listOfRoPtr": []any{clone[T](value)},
		"mapOfRoPtr": map[string]any{
			"key": clone[T](value),
		},
	}

	actual := runOnTestStruct{}
	err := model.Decode(data, &actual)

	opt := comparerForEvaluable[string]()
	expected := runOnTestStruct{
		RO:          ro,
		ROPtr:       &ro,
		ListOfRO:    []RunsOn{ro},
		MapOfRO:     map[string]RunsOn{"key": ro},
		ListOfROPtr: []*RunsOn{&ro},
		MapOfROPtr:  map[string]*RunsOn{"key": &ro},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opt)
}
