package workflows

import (
	"github.com/dungdm93/drasi/pkg/model"
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
		con := Container{Image: NewIdent("ubuntu:22.04")}
		testDecodeContainer(tt, "ubuntu:22.04", con)
	})

	t.Run("expr", func(tt *testing.T) {
		con := Container{Image: NewExpr("${{ foobar }}", toString)}
		testDecodeContainer(tt, "${{ foobar }}", con)
	})

	t.Run("map", func(tt *testing.T) {
		con := Container{
			Image: NewIdent("ubuntu:22.04"),
			Credentials: &ContainerCredentials{
				Username: NewIdent("username"),
				Password: NewExpr("${{ foobar }}", toString),
			},
		}
		testDecodeContainer(tt, map[string]any{
			"image": "ubuntu:22.04",
			"credentials": map[string]any{
				"username": "username",
				"password": "${{ foobar }}",
			},
		}, con)
	})
}

func testDecodeContainer[C any](tt *testing.T, value C, con Container) {
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

	actual := containerTestStruct{}
	err := model.Decode(data, &actual)
	assert.NilError(tt, err)

	opts := comparerForEvaluable[string]()
	expected := containerTestStruct{
		Con:          con,
		ConPtr:       &con,
		ListOfCon:    []Container{con},
		ListOfConPtr: []*Container{&con},
		MapOfCon:     map[string]Container{"key": con},
		MapOfConPtr:  map[string]*Container{"key": &con},
	}
	assert.DeepEqual(tt, actual, expected, opts...)
}
