package workflows

import (
	"github.com/dungdm93/drassi/core/pkg/model"
	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	"testing"
)

func commonComparerForEvaluable(opts ...cmp.Option) []cmp.Option {
	var o []cmp.Option
	o = append(o, comparerForEvaluable[bool](opts...)...)
	o = append(o, comparerForEvaluable[int64](opts...)...)
	o = append(o, comparerForEvaluable[float64](opts...)...)
	o = append(o, comparerForEvaluable[string](opts...)...)
	return o
}

func comparerForEvaluable[R any](opts ...cmp.Option) []cmp.Option {
	return []cmp.Option{
		comparerForIdentity[R](opts...),
		comparerForExpression[R](opts...),
	}
}

func comparerForIdentity[R any](opts ...cmp.Option) cmp.Option {
	return cmp.Comparer(func(x, y identity[R]) bool {
		return cmp.Equal(x.value, y.value, opts...)
	})
}

func comparerForExpression[R any](opts ...cmp.Option) cmp.Option {
	return cmp.Comparer(func(x, y expression[R]) bool {
		return cmp.Equal(x.expr, y.expr, opts...)
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

	opts := comparerForEvaluable[R]()
	i := NewIdent(value)
	e := NewExpr(expr, con)

	assert.DeepEqual(tt, obj.DirectValue, i, opts...)
	assert.DeepEqual(tt, obj.Expr, e, opts...)

	// List
	assert.DeepEqual(tt, obj.ListOfExpr, []Evaluable[R]{i, e}, opts...)

	//// Map
	assert.DeepEqual(tt, obj.MapOfExpr, map[string]Evaluable[R]{"first": i, "second": e}, opts...)

	//// Struct
	assert.DeepEqual(tt, obj.StructOfExpr.DirectValue, i, opts...)
	assert.DeepEqual(tt, obj.StructOfExpr.Expr, e, opts...)

	// List
	assert.DeepEqual(tt, obj.StructOfExpr.ListOfExpr, []Evaluable[R]{i, e}, opts...)

	//// Map
	assert.DeepEqual(tt, obj.StructOfExpr.MapOfExpr, map[string]Evaluable[R]{"first": i, "second": e}, opts...)
}

type conditionalTestStruct struct {
	Conditional          Conditional             `mapstructure:"conditional"`
	ConditionalPtr       *Conditional            `mapstructure:"conditionalPtr"`
	ListOfConditional    []Conditional           `mapstructure:"listOfConditional"`
	MapOfConditional     map[string]Conditional  `mapstructure:"mapOfConditional"`
	ListOfConditionalPtr []*Conditional          `mapstructure:"listOfConditionalPtr"`
	MapOfConditionalPtr  map[string]*Conditional `mapstructure:"mapOfConditionalPtr"`
}

func TestDecodeConditional(t *testing.T) {
	val := "foobar"
	con := NewConditional("foobar")
	testDecodeConditional(t, val, con)
}

func testDecodeConditional(tt *testing.T, val string, con Conditional) {
	data := map[string]any{
		"conditional":       clone(val),
		"conditionalPtr":    clone(val),
		"listOfConditional": []any{clone(val)},
		"mapOfConditional": map[string]any{
			"key": clone(val),
		},
		"listOfConditionalPtr": []string{clone(val)},
		"mapOfConditionalPtr": map[string]any{
			"key": clone(val),
		},
	}

	actual := conditionalTestStruct{}
	err := model.Decode(data, &actual)

	opt := comparerForExpression[bool]()
	expected := conditionalTestStruct{
		Conditional:          con,
		ConditionalPtr:       &con,
		ListOfConditional:    []Conditional{con},
		ListOfConditionalPtr: []*Conditional{&con},
		MapOfConditional:     map[string]Conditional{"key": con},
		MapOfConditionalPtr:  map[string]*Conditional{"key": &con},
	}
	assert.NilError(tt, err)
	assert.DeepEqual(tt, actual, expected, opt)
}
