package workflows

import (
	"drassi.run/core/pkg/model"
	"gotest.tools/v3/assert"
	"testing"
)

type concurrencyTestStruct struct {
	Con          Concurrency             `mapstructure:"con"`
	ConPtr       *Concurrency            `mapstructure:"conPtr"`
	ListOfCon    []Concurrency           `mapstructure:"listOfCon"`
	MapOfCon     map[string]Concurrency  `mapstructure:"mapOfCon"`
	ListOfConPtr []*Concurrency          `mapstructure:"listOfConPtr"`
	MapOfConPtr  map[string]*Concurrency `mapstructure:"mapOfConPtr"`
}

func TestDecodeConcurrency(t *testing.T) {
	t.Run("string", func(tt *testing.T) {
		c := Concurrency{
			Group: Evaluable[string]{Token: NewLiteralToken("group1")},
		}
		testDecodeConcurrency(tt, "group1", c)
	})

	t.Run("expr", func(tt *testing.T) {
		c := Concurrency{
			Group: Evaluable[string]{Token: NewExpressionToken("${{ foobar }}")},
		}
		testDecodeConcurrency(tt, "${{ foobar }}", c)
	})

	t.Run("map/group-expr", func(tt *testing.T) {
		c := Concurrency{
			Group:            Evaluable[string]{Token: NewExpressionToken("${{ foobar }}")},
			CancelInProgress: true,
		}
		input := map[string]any{
			"group":              "${{ foobar }}",
			"cancel-in-progress": true,
		}
		testDecodeConcurrency(tt, input, c)
	})
	t.Run("map/group-string", func(tt *testing.T) {
		c := Concurrency{
			Group:            Evaluable[string]{Token: NewLiteralToken("group1")},
			CancelInProgress: true,
		}
		input := map[string]any{
			"group":              "group1",
			"cancel-in-progress": true,
		}
		testDecodeConcurrency(tt, input, c)
	})
}

func testDecodeConcurrency(tt *testing.T, value any, con Concurrency) {
	data := map[string]any{
		"con":       value,
		"conPtr":    value,
		"listOfCon": []any{value},
		"mapOfCon": map[string]any{
			"key": value,
		},
		"listOfConPtr": []any{value},
		"mapOfConPtr": map[string]any{
			"key": value,
		},
	}

	actual := concurrencyTestStruct{}
	err := model.Decode(data, &actual)

	opts := comparerForLiteralToken()
	expected := concurrencyTestStruct{
		Con:          con,
		ConPtr:       &con,
		ListOfCon:    []Concurrency{con},
		ListOfConPtr: []*Concurrency{&con},
		MapOfCon:     map[string]Concurrency{"key": con},
		MapOfConPtr:  map[string]*Concurrency{"key": &con},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opts)
}

type permissionTestStruct struct {
	Perm Permissions `mapstructure:"perm"`
}

func TestDecodePermissions(t *testing.T) {
	t.Run("read-all", func(tt *testing.T) {
		p := Permissions{
			Actions:            PermissionsLevelRead,
			Checks:             PermissionsLevelRead,
			Contents:           PermissionsLevelRead,
			Deployments:        PermissionsLevelRead,
			Discussions:        PermissionsLevelRead,
			IdToken:            PermissionsLevelRead,
			Issues:             PermissionsLevelRead,
			Packages:           PermissionsLevelRead,
			Pages:              PermissionsLevelRead,
			PullRequests:       PermissionsLevelRead,
			RepositoryProjects: PermissionsLevelRead,
			SecurityEvents:     PermissionsLevelRead,
			Statuses:           PermissionsLevelRead,
		}
		testDecodePermissions(tt, "read-all", p)
	})

	t.Run("write-all", func(tt *testing.T) {
		p := Permissions{
			Actions:            PermissionsLevelWrite,
			Checks:             PermissionsLevelWrite,
			Contents:           PermissionsLevelWrite,
			Deployments:        PermissionsLevelWrite,
			Discussions:        PermissionsLevelWrite,
			IdToken:            PermissionsLevelWrite,
			Issues:             PermissionsLevelWrite,
			Packages:           PermissionsLevelWrite,
			Pages:              PermissionsLevelWrite,
			PullRequests:       PermissionsLevelWrite,
			RepositoryProjects: PermissionsLevelWrite,
			SecurityEvents:     PermissionsLevelWrite,
			Statuses:           PermissionsLevelWrite,
		}
		testDecodePermissions(tt, "write-all", p)
	})

	t.Run("specify", func(tt *testing.T) {
		p := Permissions{
			Actions:  PermissionsLevelWrite,
			IdToken:  PermissionsLevelRead,
			Issues:   PermissionsLevelNone,
			Statuses: PermissionsLevelWrite,
		}
		d := map[string]any{
			"actions":  PermissionsLevelWrite,
			"id-token": PermissionsLevelRead,
			"issues":   PermissionsLevelNone,
			"statuses": PermissionsLevelWrite,
		}
		testDecodePermissions(tt, d, p)
	})
}

func testDecodePermissions(tt *testing.T, value any, perm Permissions) {
	data := map[string]any{
		"perm": value,
	}

	obj := permissionTestStruct{}
	err := model.Decode(data, &obj)

	assert.NilError(tt, err)
	assert.Equal(tt, obj.Perm, perm)
}
