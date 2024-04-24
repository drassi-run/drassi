package ast

import (
	"fmt"
	"slices"

	"github.com/dungdm93/drasi/pkg/expr/ast/functions"
	"github.com/dungdm93/drasi/pkg/expr/ast/operators"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

const (
	maxExpressionLength = 21000
	maxDepth            = 50
)

type (
	parseContext struct {
		lexer                *lexer
		token                *token
		lastToken            *token
		expr                 string
		allowUnknownKeywords bool
		fnsInfo              map[string]functions.IFnInfo[interfaces.Fn]
		namedValsInfo        map[string]INamedValueInfo[INamedValue]
		operands             []interfaces.Node
		operators            []*token
	}
)

func newParseContext(expr string, namedVals []INamedValueInfo[INamedValue], fns []functions.IFnInfo[interfaces.Fn], allowUnknownKeyWords bool) *parseContext {
	result := parseContext{
		allowUnknownKeywords: allowUnknownKeyWords,
		fnsInfo:              map[string]functions.IFnInfo[interfaces.Fn]{},
		namedValsInfo:        map[string]INamedValueInfo[INamedValue]{},
	}
	if len(expr) > maxExpressionLength {
		panic(ErrorMaxLengthExceeded)
	}

	for _, value := range namedVals {
		result.namedValsInfo[value.GetName()] = value
	}

	for _, fn := range fns {
		result.fnsInfo[fn.GetName()] = fn
	}
	result.lexer = newLexer(expr)
	return &result
}

/*
CreateTree

  - Create parseContext
  - Create LexicalAnalyzer from expr
  - Add functions info to parseContext
  - Add named values info to parseContext
  - createTree with parseContext
*/

func CreateTreeWithDefaults(expr string) (astRoot interfaces.Node) {
	namedVals := []INamedValueInfo[INamedValue]{
		NewNamedValueInfo[ContextValueNode]("github"),
		NewNamedValueInfo[ContextValueNode]("strategy"),
		NewNamedValueInfo[ContextValueNode]("env"),
		NewNamedValueInfo[ContextValueNode]("steps"),
		NewNamedValueInfo[ContextValueNode]("runner"),
		NewNamedValueInfo[ContextValueNode]("strategy"),
		NewNamedValueInfo[ContextValueNode]("needs"),
		NewNamedValueInfo[ContextValueNode]("inputs"),
	}
	return createTree(newParseContext(expr, namedVals, nil, false))
}

func CreateTree(expr string, namedValues []INamedValueInfo[INamedValue],
	functions []functions.IFnInfo[interfaces.Fn]) (astRoot interfaces.Node) {
	return createTree(newParseContext(expr, namedValues, functions, false))
}

func createTree(pCtx *parseContext) interfaces.Node {
	for {
		token, haveToken := pCtx.lexer.next()
		pCtx.token = token
		if !haveToken {
			break
		}
		if pCtx.token.kind() == unexpected {
			panic(fmt.Sprintf("unexpected token, rawVal: %s, k: %v, expr: %s", pCtx.token.rawVal,
				pCtx.token.kind(),
				pCtx.expr))
		}
		if pCtx.token.isOperator() {
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
	// check unexpected end of expr
	if len(pCtx.operators) > 0 {
		var unexpectedLastToken bool
		switch pCtx.lastToken.kind() {
		// Legal
		case endGroup, endIndex, endParameters:
			break
			// Illegal
		case function:
			unexpectedLastToken = true
		default:
			unexpectedLastToken = pCtx.lastToken.isOperator()
		}
		if unexpectedLastToken || len(pCtx.lexer.getUnclosedTokens()) > 0 {
			panic(fmt.Errorf("unexpected last token, rawVal: %s, k: %v, expr: %s", pCtx.lastToken.rawVal,
				pCtx.lastToken.kind(),
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
	if pCtx.token.associativity() == associativityLTR {
		tk := pCtx.token
		for len(pCtx.operators) > 0 {
			topOp := pCtx.operators[len(pCtx.operators)-1]
			if topOp.precedence() >= tk.precedence() &&
				topOp.kind() != startGroup &&
				topOp.kind() != startIndex &&
				topOp.kind() != startParameters &&
				topOp.kind() != separator {
				flushTopOperator(pCtx)
				continue
			}
			break
		}
	}
	pCtx.operators = append(pCtx.operators, pCtx.token)
	// Process closing operators now, since context.lastToken is required
	// to accurately process TokenKind.endParameters
	switch pCtx.token.kind() {
	case endGroup, endIndex, endParameters:
		flushTopOperator(pCtx)
	}
}

func pushOperand(pCtx *parseContext) {
	switch pCtx.token.kind() {
	case function:
		fn := pCtx.token.rawVal
		if fnInfo := tryGetFnInfo(pCtx, fn); fnInfo != nil {
			node := fnInfo.CreateNode().(interfaces.Fn).(interfaces.Node)
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
	case namedValue:
		name := pCtx.token.rawVal
		if namedValInfo, exist := pCtx.namedValsInfo[name]; exist {
			node := namedValInfo.CreateNode().(INamedValue).(interfaces.Node)
			node.SetName(name)
			pCtx.operands = append(pCtx.operands, node)
		} else {
			if pCtx.allowUnknownKeywords {
				node := new(noOpNamedValue)
				node.SetName(name)
				pCtx.operands = append(pCtx.operands, node)
			} else {
				panic(fmt.Errorf("unrecognized named value"))
			}
		}

	default:
		pCtx.operands = append(pCtx.operands, pCtx.token.toNode())
	}
}

func flushTopOperator(pCtx *parseContext) {
	switch pCtx.operators[len(pCtx.operators)-1].kind() {
	case endIndex: // "]"
		flushTopEndIndex(pCtx)
		return
	case endGroup: // ")" logical grouping
		flushTopEndGroup(pCtx)
		return
	case endParameters: // ")" function call
		flushTopEndParameters(pCtx)
		return
	}
	// remove top operator
	tk := pCtx.operators[len(pCtx.operators)-1]
	pCtx.operators = pCtx.operators[:len(pCtx.operators)-1]

	node := tk.toNode().(interfaces.Container)
	operands := popOperands(pCtx, tk.operandCnt())
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
func strictPopOnOperator(pCtx *parseContext, expectedKind tokenKind) (popped *token) {
	top := pCtx.operators[len(pCtx.operators)-1]
	if top.kind() != expectedKind {
		panic(fmt.Sprintf("expected operator %v to be of k %v", expectedKind, top.kind()))
	}
	pCtx.operators = pCtx.operators[:len(pCtx.operators)-1]
	return top
}

// popOperands remove count element from the operands stack
func popOperands(pCtx *parseContext, count int) []interfaces.Node {
	var result []interfaces.Node
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
	strictPopOnOperator(pCtx, endIndex)
	// Pop start pos
	tk := strictPopOnOperator(pCtx, startIndex)
	node := tk.toNode().(interfaces.Container)

	ops := popOperands(pCtx, tk.operandCnt())
	for _, o := range ops {
		node.AddParameter(o)
	}
	pCtx.operands = append(pCtx.operands, node)
}

// flushTopEndGroup remove top logical group ")" from operator stack
func flushTopEndGroup(pCtx *parseContext) {
	strictPopOnOperator(pCtx, endGroup)
	strictPopOnOperator(pCtx, startGroup)
}

// flushTopEndParameters remove top end parameter ")" end of function call
func flushTopEndParameters(pCtx *parseContext) {
	tk := strictPopOnOperator(pCtx, endParameters)
	// Sanity check top tk is the current token
	if tk != pCtx.token {
		panic(fmt.Errorf("expected popped token to be the current token"))
	}
	var fn interfaces.Fn
	// no parameter fn
	if pCtx.lastToken.kind() == startParameters {
		// node already exist on operand stack
		fn = pCtx.operands[len(pCtx.operands)-1].(interfaces.Fn)
	} else {
		// parameter fn
		var seperatorCnt int
		for pCtx.operators[len(pCtx.operators)-1].kind() == separator {
			seperatorCnt++
			pCtx.operators = pCtx.operators[:len(pCtx.operators)-1]
		}
		// eg: func(x,y,z) -> 2 separator -> 3 params
		fnOperands := popOperands(pCtx, seperatorCnt+1)
		// node already exist on operand stack
		fn = pCtx.operands[len(pCtx.operands)-1].(interfaces.Fn)
		// add operand to the fn node
		for _, operand := range fnOperands {
			fn.AddParameter(operand)
		}
	}
	strictPopOnOperator(pCtx, startParameters)
	fnInfo := tryGetFnInfo(pCtx, fn.GetName())
	if fnInfo != nil && pCtx.allowUnknownKeywords {
	}
	if err := fnLimitCheck(fn, fnInfo); err != nil {
		panic(err)
	}
}

// update to return type maybe ?
func tryGetFnInfo(pCtx *parseContext, name string) (result functions.IFnInfo[interfaces.Fn]) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		result = nil
	// 	}
	// }()
	eFn, existEfn := newFnStore().WellKnownFns[name]
	cFn, existCfn := pCtx.fnsInfo[name]
	if existEfn && eFn.GetName() != "" {
		result = eFn
	}
	// if eFn.GetName() != "" {
	// 	result = eFn
	// }
	if existCfn {
		result = cFn
	}
	return result
}

func fnLimitCheck(f interfaces.Fn, expected functions.IFnInfo[interfaces.Fn]) (err error) {
	if len(f.Parameters()) < expected.MinParameters() {
		err = ErrorTooFewParameters
	}
	if len(f.Parameters()) > expected.MaxParameters() {
		err = ErrorTooManyParameters
	}
	return err
}

func checkMaxDepth(pCtx *parseContext, node interfaces.Node, depth int) (err error) {
	if depth > maxDepth {
		return ErrorMaxDepthExceeded
	}
	if container, isContainer := node.(interfaces.Container); isContainer {
		for _, param := range container.Parameters() {
			_ = checkMaxDepth(pCtx, param, depth+1)
		}
	}
	return nil
}
