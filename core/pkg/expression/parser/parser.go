package parser

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expression/ast"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/functions"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/operators"
	"github.com/dungdm93/drassi/core/pkg/expression/scanner"
	token2 "github.com/dungdm93/drassi/core/pkg/expression/token"
)

const (
	maxExpressionLength = 21000
	maxDepth            = 50
)

type (
	parseContext struct {
		scanner              *scanner.Scanner
		token                *token2.Token
		lastToken            *token2.Token
		expr                 string
		allowUnknownKeywords bool
		fnsInfo              map[string]functions.IFnInfo[ast_ifaces.Fn]
		namedValsInfo        map[string]ast.INamedValueInfo[ast.INamedValue]
		operands             []ast_ifaces.ExprNode
		operators            []*token2.Token
	}
)

var (
	ErrorUnrecognizedNamedValue = fmt.Errorf("unrecognized named value")
)

func newParseContext(expr string, namedVals []ast.INamedValueInfo[ast.INamedValue], fns []functions.IFnInfo[ast_ifaces.Fn], allowUnknownKeyWords bool) *parseContext {
	result := parseContext{
		allowUnknownKeywords: allowUnknownKeyWords,
		fnsInfo:              map[string]functions.IFnInfo[ast_ifaces.Fn]{},
		namedValsInfo:        map[string]ast.INamedValueInfo[ast.INamedValue]{},
	}
	if len(expr) > maxExpressionLength {
		panic(ast.ErrorMaxLengthExceeded)
	}

	for _, value := range namedVals {
		result.namedValsInfo[value.GetName()] = value
	}

	for _, fn := range fns {
		result.fnsInfo[fn.GetName()] = fn
	}
	result.scanner = scanner.NewScanner(expr)
	return &result
}

/*
Parse

  - Create parseContext
  - Create Scanner for expression
  - Add functions info to parseContext
  - Add named values info to parseContext
  - parse with parseContext
*/

func ParseWithDefaults(expr string) (root ast_ifaces.ExprNode) {
	namedVals := []ast.INamedValueInfo[ast.INamedValue]{
		ast.NewNamedValueInfo[ast.ContextValueNode]("github"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("strategy"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("env"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("steps"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("runner"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("strategy"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("needs"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("inputs"),
	}
	return parse(newParseContext(expr, namedVals, nil, false))
}

func Parse(expr string, namedValues []ast.INamedValueInfo[ast.INamedValue],
	functions []functions.IFnInfo[ast_ifaces.Fn]) (root ast_ifaces.ExprNode) {
	return parse(newParseContext(expr, namedValues, functions, false))
}

func parse(pCtx *parseContext) ast_ifaces.ExprNode {
	for {
		token, haveToken := pCtx.scanner.Next()
		pCtx.token = token
		if !haveToken {
			break
		}
		if pCtx.token.Kind() == token2.Unexpected {
			panic(fmt.Sprintf("unexpected token, rawVal: %s, k: %v, expression: %s", pCtx.token.RawVal,
				pCtx.token.Kind(),
				pCtx.expr))
		}
		if pCtx.token.IsOperator() {
			pushOperator(pCtx)
		} else {
			pushOperand(pCtx)
		}
		pCtx.lastToken = pCtx.token
	}
	// no tokens
	if pCtx.lastToken == nil {
		return nil
	}
	// check unexpected end of expression
	if len(pCtx.operators) > 0 {
		var unexpectedLastToken bool
		switch pCtx.lastToken.Kind() {
		// Legal
		case token2.EndGroup, token2.EndIndex, token2.EndParameters:
			break
			// Illegal
		case token2.Function:
			unexpectedLastToken = true
		default:
			unexpectedLastToken = pCtx.lastToken.IsOperator()
		}
		if unexpectedLastToken || len(pCtx.scanner.GetUnclosedTokens()) > 0 {
			panic(fmt.Errorf("unexpected last token, rawVal: %s, k: %v, expression: %s", pCtx.lastToken.RawVal,
				pCtx.lastToken.Kind(),
				pCtx.expr))
		}
	}
	for len(pCtx.operators) > 0 {
		flushTopOperator(pCtx)
	}
	if len(pCtx.operands) > 1 {
		panic("invalid number of operands")
	}
	root := pCtx.operands[0]
	if err := checkMaxDepth(pCtx, root, 1); err != nil {
		panic(err)
	}
	return root
}

func pushOperator(pCtx *parseContext) {
	if pCtx.token.Associativity() == token2.AssociativityLTR {
		tk := pCtx.token
		for len(pCtx.operators) > 0 {
			topOp := pCtx.operators[len(pCtx.operators)-1]
			if topOp.Precedence() >= tk.Precedence() &&
				topOp.Kind() != token2.StartGroup &&
				topOp.Kind() != token2.StartIndex &&
				topOp.Kind() != token2.StartParameters &&
				topOp.Kind() != token2.Separator {
				flushTopOperator(pCtx)
				continue
			}
			break
		}
	}
	pCtx.operators = append(pCtx.operators, pCtx.token)
	// Process closing operators now, since context.lastToken is required
	// to accurately process TokenKind.endParameters
	switch pCtx.token.Kind() {
	case token2.EndGroup, token2.EndIndex, token2.EndParameters:
		flushTopOperator(pCtx)
	}
}

func pushOperand(pCtx *parseContext) {
	switch pCtx.token.Kind() {
	case token2.Function:
		fn := pCtx.token.RawVal
		if fnInfo := tryGetFnInfo(pCtx, fn); fnInfo != nil {
			node := fnInfo.CreateNode().(ast_ifaces.Fn).(ast_ifaces.ExprNode)
			node.SetName(fn)
			pCtx.operands = append(pCtx.operands, node)
		} else {
			if pCtx.allowUnknownKeywords {
				node := new(functions.NoOp)
				node.SetName(fn)
				pCtx.operands = append(pCtx.operands, node)
			} else {
				panic(fmt.Errorf("unrecognized function"))
			}
		}
	case token2.NamedValue:
		name := pCtx.token.RawVal
		if namedValInfo, exist := pCtx.namedValsInfo[name]; exist {
			node := namedValInfo.CreateNode().(ast.INamedValue).(ast_ifaces.ExprNode)
			node.SetName(name)
			pCtx.operands = append(pCtx.operands, node)
		} else {
			if pCtx.allowUnknownKeywords {
				node := new(ast.NoOpNamedValue)
				node.SetName(name)
				pCtx.operands = append(pCtx.operands, node)
			} else {
				panic(ErrorUnrecognizedNamedValue)
			}
		}

	default:
		pCtx.operands = append(pCtx.operands, ast.ToNode(pCtx.token))
	}
}

func flushTopOperator(pCtx *parseContext) {
	switch pCtx.operators[len(pCtx.operators)-1].Kind() {
	case token2.EndIndex: // "]"
		flushTopEndIndex(pCtx)
		return
	case token2.EndGroup: // ")" logical grouping
		flushTopEndGroup(pCtx)
		return
	case token2.EndParameters: // ")" function call
		flushTopEndParameters(pCtx)
		return
	}
	// remove top operator
	tk := pCtx.operators[len(pCtx.operators)-1]
	pCtx.operators = pCtx.operators[:len(pCtx.operators)-1]

	node := ast.ToNode(tk).(ast_ifaces.Container)
	operands := popOperands(pCtx, tk.OperandCnt())
	_, isAnd := node.(*operators.And)
	_, isOr := node.(*operators.Or)
	for _, o := range operands {
		if isAnd {
			op, nestedArn := o.(*operators.And)
			if nestedArn {
				for _, p := range op.Parameters() {
					node.AddParameter(p)
				}
				continue
			}
		}
		if isOr {
			op, nestedOr := o.(*operators.Or)
			if nestedOr {
				for _, p := range op.Parameters() {
					node.AddParameter(p)
				}
				continue
			}
		}
		node.AddParameter(o)
	}
	// Push the node to the operand stack
	pCtx.operands = append(pCtx.operands, node)
}

// strictPopOnOperator remove top operator from the operators stack and verify kind is expectedKind,
// panic when mismatch kind
func strictPopOnOperator(pCtx *parseContext, expectedKind token2.Kind) (popped *token2.Token) {
	top := pCtx.operators[len(pCtx.operators)-1]
	if top.Kind() != expectedKind {
		panic(fmt.Sprintf("expected operator %v to be of k %v", expectedKind, top.Kind()))
	}
	pCtx.operators = pCtx.operators[:len(pCtx.operators)-1]
	return top
}

// popOperands remove count element from the operands stack
func popOperands(pCtx *parseContext, count int) []ast_ifaces.ExprNode {
	var result []ast_ifaces.ExprNode
	for i := 0; i < count; i++ {
		result = append(result, pCtx.operands[len(pCtx.operands)-1])
		pCtx.operands = pCtx.operands[:len(pCtx.operands)-1]
	}
	slices.Reverse(result)
	return result
}

// flushTopEndIndex:
// - remove top end index from operator stack,
// - calculate its required operands
// - create a new container node with parameters point to operator node as parent
// - push back container node to operands stack
func flushTopEndIndex(pCtx *parseContext) {
	// Pop the operators
	// Pop end pos
	strictPopOnOperator(pCtx, token2.EndIndex)
	// Pop start pos
	tk := strictPopOnOperator(pCtx, token2.StartIndex)
	node := ast.ToNode(tk).(ast_ifaces.Container)

	ops := popOperands(pCtx, tk.OperandCnt())
	for _, o := range ops {
		node.AddParameter(o)
	}
	pCtx.operands = append(pCtx.operands, node)
}

// flushTopEndGroup remove top logical group ")" from operator stack
func flushTopEndGroup(pCtx *parseContext) {
	strictPopOnOperator(pCtx, token2.EndGroup)
	strictPopOnOperator(pCtx, token2.StartGroup)
}

// flushTopEndParameters remove top end parameter ")" end of function call
func flushTopEndParameters(pCtx *parseContext) {
	tk := strictPopOnOperator(pCtx, token2.EndParameters)
	// Sanity check top tk is the current token
	if tk != pCtx.token {
		panic(fmt.Errorf("expected popped token to be the current token"))
	}
	var fn ast_ifaces.Fn
	// no parameter fn
	if pCtx.lastToken.Kind() == token2.StartParameters {
		// node already exist on operand stack
		fn = pCtx.operands[len(pCtx.operands)-1].(ast_ifaces.Fn)
	} else {
		// parameter fn
		var seperatorCnt int
		for pCtx.operators[len(pCtx.operators)-1].Kind() == token2.Separator {
			seperatorCnt++
			pCtx.operators = pCtx.operators[:len(pCtx.operators)-1]
		}
		// eg: func(x,y,z) -> 2 separator -> 3 params
		fnOperands := popOperands(pCtx, seperatorCnt+1)
		// node already exist on operand stack
		fn = pCtx.operands[len(pCtx.operands)-1].(ast_ifaces.Fn)
		// add operand to the fn node
		for _, operand := range fnOperands {
			fn.AddParameter(operand)
		}
	}
	strictPopOnOperator(pCtx, token2.StartParameters)
	fnInfo := tryGetFnInfo(pCtx, fn.GetName())
	if fnInfo != nil && pCtx.allowUnknownKeywords {
	}
	if err := fnLimitCheck(fn, fnInfo); err != nil {
		panic(err)
	}
}

func tryGetFnInfo(pCtx *parseContext, name string) (result functions.IFnInfo[ast_ifaces.Fn]) {
	eFn, existEfn := ast.NewFnStore().WellKnownFns[strings.ToLower(name)]
	cFn, existCfn := pCtx.fnsInfo[strings.ToLower(name)]
	if existEfn && eFn.GetName() != "" {
		result = eFn
	}
	if existCfn {
		result = cFn
	}
	return result
}

func fnLimitCheck(f ast_ifaces.Fn, expected functions.IFnInfo[ast_ifaces.Fn]) (err error) {
	if len(f.Parameters()) < expected.MinParameters() {
		err = ast.ErrorTooFewParameters
	}
	if len(f.Parameters()) > expected.MaxParameters() {
		err = ast.ErrorTooManyParameters
	}
	return err
}

func checkMaxDepth(pCtx *parseContext, node ast_ifaces.ExprNode, depth int) (err error) {
	if depth > maxDepth {
		return ast.ErrorMaxDepthExceeded
	}
	if container, isContainer := node.(ast_ifaces.Container); isContainer {
		for _, param := range container.Parameters() {
			_ = checkMaxDepth(pCtx, param, depth+1)
		}
	}
	return nil
}
