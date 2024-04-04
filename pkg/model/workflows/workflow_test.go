package workflows

import (
	"github.com/dungdm93/drasi/pkg/model"
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
			Group: newIdent("group1"),
		}
		testDecodeConcurrency(tt, "group1", c)
	})

	t.Run("expr", func(tt *testing.T) {
		c := Concurrency{
			Group: newExpr("${{ foobar }}", toString),
		}
		testDecodeConcurrency(tt, "${{ foobar }}", c)
	})

	t.Run("map", func(tt *testing.T) {
		c := Concurrency{
			Group:            newExpr("${{ foobar }}", toString),
			CancelInProgress: true,
		}
		testDecodeConcurrency(tt, map[string]any{
			"group":              "${{ foobar }}",
			"cancel-in-progress": true,
		}, c)
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

	obj := concurrencyTestStruct{}
	err := model.Decode(data, &obj)

	assert.NilError(tt, err)
	assert.Check(tt, assertEqualConcurrency(obj.Con, con))
	assert.Check(tt, assertEqualConcurrency(*obj.ConPtr, con))
	assert.Equal(tt, len(obj.ListOfCon), 1)
	assert.Check(tt, assertEqualConcurrency(obj.ListOfCon[0], con))
	assert.Equal(tt, len(obj.MapOfCon), 1)
	assert.Check(tt, assertEqualConcurrency(obj.MapOfCon["key"], con))
	assert.Equal(tt, len(obj.ListOfConPtr), 1)
	assert.Check(tt, assertEqualConcurrency(*obj.ListOfConPtr[0], con))
	assert.Equal(tt, len(obj.MapOfConPtr), 1)
	assert.Check(tt, assertEqualConcurrency(*obj.MapOfConPtr["key"], con))
}

func assertEqualConcurrency(a, b Concurrency) bool {
	if a.CancelInProgress != b.CancelInProgress {
		return false
	}
	if ai, aok := a.Group.(identity[string]); aok {
		if bi, bok := b.Group.(identity[string]); bok {
			return ai.value == bi.value
		}
		return false
	}
	if ae, aok := a.Group.(expression[string]); aok {
		if be, bok := b.Group.(expression[string]); bok {
			return ae.expr == be.expr
		}
		return false
	}
	return false
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
