package workflows

import (
	"github.com/mitchellh/mapstructure"
	"gotest.tools/v3/assert"
	"testing"
)

type containerTestStruct struct {
	Con          Container             `mapstructure:"con"`
	ConPtr       *Container            `mapstructure:"conPtr"`
	ListOfCon    []Container           `mapstructure:"listOfCon"`
	MapOfCon     map[string]Container  `mapstructure:"mapOfCon"`
	ListOfConPtr []*Container          `mapstructure:"listOfConPtr"`
	MapOfConPtr  map[string]*Container `mapstructure:"mapOfConPtr"`
}

func TestDecodeContainer(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		testDecodeContainer[string](tt, "ubuntu:22.04")
	})

	t.Run("expr", func(tt *testing.T) {
		testDecodeContainer[string](tt, "${{ foobar }}")
	})

	t.Run("map", func(tt *testing.T) {
		testDecodeContainer[map[string]any](tt, map[string]any{
			"image": "ubuntu:22.04",
			"credentials": map[string]any{
				"username": "username",
				"password": "${{ foobar }}",
			},
		})
	})
}

func testDecodeContainer[C any](tt *testing.T, value C) {
	data := map[string]any{
		"con":       value,
		"conPtr":    value,
		"listOfCon": []C{value},
		"mapOfCon": map[string]any{
			"key": value,
		},
		"listOfConPtr": []C{value},
		"mapOfConPtr": map[string]any{
			"key": value,
		},
	}

	obj := containerTestStruct{}
	err := decode(data, &obj)

	assert.NilError(tt, err)
}

func decode(source any, target any) error {
	metadata := mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			DecoderHook,
			DecodeEvaluableHook,
		),
		Result:   target,
		TagName:  "mapstructure",
		Metadata: &metadata,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(source)
}
