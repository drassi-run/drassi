package model

import (
	"github.com/mitchellh/mapstructure"
	"gotest.tools/v3/assert"
	"reflect"
	"testing"
)

type klass[E any] struct {
	DirectValue  E            `json:"direct,omitempty" yaml:"direct,omitempty"`
	Expr         E            `json:"expr,omitempty" yaml:"expr,omitempty"`
	ListOfExpr   []E          `json:"list_of_expr,omitempty" yaml:"list_of_expr,omitempty"`
	MapOfExpr    map[string]E `json:"map_of_expr,omitempty" yaml:"map_of_expr,omitempty"`
	StructOfExpr struct {
		DirectValue E            `json:"direct,omitempty" yaml:"direct,omitempty"`
		Expr        E            `json:"expr,omitempty" yaml:"expr,omitempty"`
		ListOfExpr  []E          `json:"list_of_expr,omitempty" yaml:"list_of_expr,omitempty"`
		MapOfExpr   map[string]E `json:"map_of_expr,omitempty" yaml:"map_of_expr,omitempty"`
	} `json:"struct_of_expr,omitempty" yaml:"struct_of_expr,omitempty"`
}

func TestDecodeEvaluableHook(t *testing.T) {
	t.Run("bool", func(tt *testing.T) {
		testDecodeEvaluableHook[bool](tt, true)
	})

	t.Run("int64", func(tt *testing.T) {
		testDecodeEvaluableHook[int64](tt, 123456)
	})

	t.Run("float64", func(tt *testing.T) {
		testDecodeEvaluableHook[float64](tt, 123456.789)
	})

	t.Run("string", func(tt *testing.T) {
		testDecodeEvaluableHook[string](tt, "hello world")
	})
}

func testDecodeEvaluableHook[R any](t *testing.T, value R) {
	var expr = "${{ expr }}"
	var list = []any{value, "${{ expr }}"}
	var dict = map[string]any{
		"first":  value,
		"second": "${{ expr }}",
	}
	var strct = map[string]any{
		"direct":       value,
		"expr":         expr,
		"list_of_expr": list,
		"map_of_expr":  dict,
	}
	var data = map[string]any{
		"direct":         value,
		"expr":           expr,
		"list_of_expr":   list,
		"map_of_expr":    dict,
		"struct_of_expr": strct,
	}
	obj := klass[Evaluable[R]]{}
	err := decode(data, &obj)

	assert.NilError(t, err)

	assert.Equal(t, obj.DirectValue, newIdent(value))
	compareExpr(t, obj.Expr, expr)

	// List
	assert.Equal(t, len(obj.ListOfExpr), 2)
	assert.Equal(t, obj.ListOfExpr[0], newIdent(value))
	compareExpr(t, obj.ListOfExpr[1], expr)

	// Map
	assert.Equal(t, len(obj.MapOfExpr), 2)
	assert.Equal(t, obj.MapOfExpr["first"], newIdent(value))
	compareExpr(t, obj.MapOfExpr["second"], expr)

	// Struct
	assert.Equal(t, obj.StructOfExpr.DirectValue, newIdent(value))
	compareExpr(t, obj.StructOfExpr.Expr, expr)

	assert.Equal(t, len(obj.StructOfExpr.ListOfExpr), 2)
	assert.Equal(t, obj.StructOfExpr.ListOfExpr[0], newIdent(value))
	compareExpr(t, obj.StructOfExpr.ListOfExpr[1], expr)

	assert.Equal(t, len(obj.StructOfExpr.MapOfExpr), 2)
	assert.Equal(t, obj.StructOfExpr.MapOfExpr["first"], newIdent(value))
	compareExpr(t, obj.StructOfExpr.MapOfExpr["second"], expr)
}

func compareExpr[R any](t *testing.T, obj Evaluable[R], expr string) {
	assert.Equal(t, reflect.TypeOf(obj), reflect.TypeFor[expression[R]]())
	var e = obj.(expression[R])
	assert.Equal(t, e.expr, expr)
}

func decode(source any, target any) error {
	metadata := mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		DecodeHook: DecodeEvaluableHook,
		Result:     target,
		TagName:    "yaml",
		Metadata:   &metadata,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(source)
}
