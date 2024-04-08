package workflows

import (
	"github.com/dungdm93/drasi/pkg/model"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	"reflect"
	"testing"
)

func comparerForEvaluable[R any](opts ...cmp.Option) cmp.Option {
	return cmp.Comparer(func(x, y Evaluable[R]) bool {
		if x == nil {
			return y == nil
		}
		if y == nil {
			return x == nil
		}

		// x != nil && y != nil
		if reflect.TypeOf(x) != reflect.TypeOf(y) {
			return false
		}
		switch a := x.(type) {
		case identity[R]:
			b, ok := y.(identity[R])
			return ok && cmp.Equal(a.value, b.value, opts...)
		case expression[R]:
			b, ok := y.(expression[R])
			return ok && cmp.Equal(a.expr, b.expr, opts...)
		}
		return false
	})
}

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
		testDecodeEvaluableHook[bool](tt, true, toBool)
	})

	t.Run("int64", func(tt *testing.T) {
		testDecodeEvaluableHook[int64](tt, 123456, toInteger)
	})

	t.Run("float64", func(tt *testing.T) {
		testDecodeEvaluableHook[float64](tt, 123456.789, toFloat)
	})

	t.Run("string", func(tt *testing.T) {
		testDecodeEvaluableHook[string](tt, "hello world", toString)
	})
}

func testDecodeEvaluableHook[R any](tt *testing.T, value R, con converter[R]) {
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

	assert.NilError(tt, err)

	opt := comparerForEvaluable[R]()
	i := NewIdent(value)
	e := NewExpr(expr, con)

	assert.DeepEqual(tt, obj.DirectValue, i, opt)
	assert.DeepEqual(tt, obj.Expr, e, opt)

	// List
	assert.DeepEqual(tt, obj.ListOfExpr, []Evaluable[R]{i, e}, opt)

	//// Map
	assert.DeepEqual(tt, obj.MapOfExpr, map[string]Evaluable[R]{"first": i, "second": e}, opt)

	//// Struct
	assert.DeepEqual(tt, obj.StructOfExpr.DirectValue, i, opt)
	assert.DeepEqual(tt, obj.StructOfExpr.Expr, e, opt)

	// List
	assert.DeepEqual(tt, obj.StructOfExpr.ListOfExpr, []Evaluable[R]{i, e}, opt)

	//// Map
	assert.DeepEqual(tt, obj.StructOfExpr.MapOfExpr, map[string]Evaluable[R]{"first": i, "second": e}, opt)
}
