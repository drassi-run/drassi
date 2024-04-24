package workflows

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	"github.com/dungdm93/drasi/pkg/model"
	"github.com/dungdm93/drasi/pkg/model/contexts"
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

	// // Map
	assert.DeepEqual(tt, obj.MapOfExpr, map[string]Evaluable[R]{"first": i, "second": e}, opts...)

	// // Struct
	assert.DeepEqual(tt, obj.StructOfExpr.DirectValue, i, opts...)
	assert.DeepEqual(tt, obj.StructOfExpr.Expr, e, opts...)

	// List
	assert.DeepEqual(tt, obj.StructOfExpr.ListOfExpr, []Evaluable[R]{i, e}, opts...)

	// // Map
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

func TestEvaluateExpr(t *testing.T) {
	t.Run("literal", func(tt *testing.T) {
		testEvaluateExpr[bool](tt, "true", true, toBool, contexts.Context{})
		testEvaluateExpr[float64](tt, "123456", 123456, toFloat, contexts.Context{})
		testEvaluateExpr[float64](tt, "123456.789", 123456.789, toFloat, contexts.Context{})
		testEvaluateExpr[string](tt, "'hello world'", "hello world", toString, contexts.Context{})
	})
	t.Run("logical", func(tt *testing.T) {
		testEvaluateExpr[bool](tt, "true", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "!true && false", false, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "1 == 1", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "(1 == 1) && 2 == 2", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "false || true", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "1 < 2", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "1 != 1", false, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "(3 <= 3) || (4 > 5)", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "!((3 > 3) && (4 >= 4))", true, toBool, contexts.Context{})
	})
	t.Run("string fmt", func(tt *testing.T) {
		testEvaluateExpr[bool](tt, "contains('Hello world', 'llo')", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "startsWith('Hello world', 'He')", true, toBool, contexts.Context{})
		testEvaluateExpr[bool](tt, "endsWith('Hello world', 'world')", true, toBool, contexts.Context{})
		testEvaluateExpr[string](tt, "format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat", toString, contexts.Context{})
		testEvaluateExpr[string](tt, "format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}", toString, contexts.Context{})
		testEvaluateExpr[string](tt, "format('Result: {0}', 1 > 2 && 3 > 4)", "Result: false", toString, contexts.Context{})
		testEvaluateExpr[string](tt, "format('Result: {0}', 1 > 2 || 3 < 4)", "Result: true", toString, contexts.Context{})
	})
	t.Run("access context data", func(tt *testing.T) {
		testEvaluateExpr[string](tt, "github.actor", "foo", toString, contexts.Context{Github: contexts.Github{Actor: "foo"}})
		testEvaluateExpr[string](tt, "format('github.actor: {0}', github.actor)", "github.actor: foo", toString, contexts.Context{Github: contexts.Github{Actor: "foo"}})
	})
}

func testEvaluateExpr[R any](tt *testing.T, expr string, expected R, con converter[R], ctx contexts.Context) {
	e := NewExpr(expr, con)
	res, err := e.Evaluate(contexts.GoContext(context.Background(), ctx))
	assert.NilError(tt, err)
	opts := comparerForEvaluable[R]()
	assert.DeepEqual(tt, res, expected, opts...)
}
