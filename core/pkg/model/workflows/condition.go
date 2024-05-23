package workflows

import (
	"fmt"
)

// Conditional is Evaluable[bool] type used by `if`, `pre-if` and `post-if`.
// The `${{ }}` expression syntax is optional and can be omitted. GitHub Actions always evaluates it as an expression.
type Conditional interface {
	Meet(name string, provider EvaluatorProvider) (bool, error)
}

type cond string

func (c *cond) Meet(name string, provider EvaluatorProvider) (bool, error) {
	ctx := provider.ContextData(name)
	if b, ok := ctx.Value(c).(bool); ok { // TODO real expression evaluation
		return b, nil
	} else {
		return false, fmt.Errorf("%s is not a bool", string(*c))
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
