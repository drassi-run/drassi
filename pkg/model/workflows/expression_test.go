package workflows

import (
	"github.com/dungdm93/drasi/pkg/model"
	"gotest.tools/v3/assert"
	"reflect"
	"testing"
)

type evaluableTestStruct[E any] struct {
	DirectValue  E            `mapstructure:"direct,omitempty"`
	Expr         E            `mapstructure:"expr,omitempty"`
	ListOfExpr   []E          `mapstructure:"list_of_expr,omitempty"`
	MapOfExpr    map[string]E `mapstructure:"map_of_expr,omitempty"`
	StructOfExpr struct {
		DirectValue E            `mapstructure:"direct,omitempty"`
		Expr        E            `mapstructure:"expr,omitempty"`
		ListOfExpr  []E          `mapstructure:"list_of_expr,omitempty"`
		MapOfExpr   map[string]E `mapstructure:"map_of_expr,omitempty"`
	} `mapstructure:"struct_of_expr,omitempty"`
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
	obj := evaluableTestStruct[Evaluable[R]]{}
	err := model.Decode(data, &obj)

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
