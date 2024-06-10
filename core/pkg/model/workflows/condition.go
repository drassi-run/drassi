package workflows

import (
	"fmt"
	"math"

	"github.com/dungdm93/drassi/core/pkg/expression/ast"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/functions"
	"github.com/dungdm93/drassi/core/pkg/expression/evaluator"
	"github.com/dungdm93/drassi/core/pkg/expression/parser"
	"github.com/dungdm93/drassi/core/pkg/model/contexts"
)

// Conditional is Evaluable[bool] type used by `if`, `pre-if` and `post-if`.
// The `${{ }}` expression syntax is optional and can be omitted. GitHub Actions always evaluates it as an expression.
type Conditional interface {
	Meet(name string, provider EvaluatorProvider) (bool, error)
}

type cond string

func (c *cond) Meet(name string, provider EvaluatorProvider) (bool, error) {
	ctx := provider.ContextData(name)
	keys := provider.Functions(name)
	var availableFuncs []functions.IFnInfo[ast_ifaces.Fn]
	for _, k := range keys {
		switch k {
		case "always":
			availableFuncs = append(availableFuncs, functions.NewFunctionInfo[functions.Always]("always", 0, math.MaxInt32))
		case "cancelled":
			availableFuncs = append(availableFuncs, functions.NewFunctionInfo[functions.Cancelled]("cancelled", 0, 0))
		case "success":
			availableFuncs = append(availableFuncs, functions.NewFunctionInfo[functions.Success]("success", 0, 0))
		case "failure":
			availableFuncs = append(availableFuncs, functions.NewFunctionInfo[functions.Failure]("failure", 0, 0))
		case "hashfile":
			availableFuncs = append(availableFuncs, functions.NewFunctionInfo[functions.HashFile]("hashfile", 1, math.MaxUint8))
		default:
		}
	}
	availableContexts := []ast.INamedValueInfo[ast.INamedValue]{
		ast.NewNamedValueInfo[ast.ContextValueNode]("github"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("strategy"),
	}
	a := parser.Parse(string(*c), availableContexts, availableFuncs)
	// TODO: proper way to set pass working dir
	r, err  := evaluator.EvaluateWithDefaults(a, &contexts.Expr{State: &ctx}, "")
	if err != nil {
		return false, err
	}
	if b, ok := r.Value().(bool); ok {
		return b, nil
	} else {
		return false, fmt.Errorf("%s evaluated to %#v is not a bool", string(*c), r.Value())
	}
}

func NewConditional(s string) Conditional {
	c := cond(s)
	return &c
}

type biCond struct {
	left  Conditional
	right Conditional
}

type orCond biCond

func (c *orCond) Meet(name string, provider EvaluatorProvider) (bool, error) {
	l, err := c.left.Meet(name, provider)
	if err != nil || !l {
		return l, err
	}

	r, err := c.right.Meet(name, provider)
	if err != nil || !r {
		return r, err
	}

	return true, nil
}

func NewConditionalOr(l, r Conditional) Conditional {
	return &orCond{left: l, right: r}
}

type andCond biCond

func (c *andCond) Meet(name string, provider EvaluatorProvider) (bool, error) {
	l, err := c.left.Meet(name, provider)
	if err != nil || l {
		return l, err
	}

	r, err := c.right.Meet(name, provider)
	if err != nil || r {
		return r, err
	}

	return false, nil
}

func NewConditionalAnd(l, r Conditional) Conditional {
	return &andCond{left: l, right: r}
}

type notCond struct {
	original Conditional
}

func (c *notCond) Meet(name string, provider EvaluatorProvider) (bool, error) {
	o, err := c.original.Meet(name, provider)
	if err != nil {
		return o, err
	} else {
		return !o, nil
	}
}

func NewConditionalNot(c Conditional) Conditional {
	return &notCond{original: c}
}
