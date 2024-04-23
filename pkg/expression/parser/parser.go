package parser

import (
	"errors"
	"fmt"
	"slices"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/functions"
	"github.com/dungdm93/drasi/pkg/expression/parser/operators"
)

const (
	maxExpressionLength = 21000
)

var (
	ErrorsTooFewParameters  = errors.New("too few parameter")
	ErrorsTooManyParameters = errors.New("too many parameter")
	ErrorsMaxDepthExceeded  = errors.New("exceeded max depth")
	ErrorsMaxLengthExceeded = errors.New("exceed max length")
)

type (
	parseContext struct {
		Lexer                *lexer
		Token                *lexicalToken
		LastToken            *lexicalToken
		Expression           string
		AllowUnknownKeywords bool
		FnsInfo              map[string]functions.IFnInfo[functions.IFn]
		NamedValsInfo        map[string]INamedValueInfo[INamedValue]
		Operands             []expression.IExpNode
		Operators            []*lexicalToken
	}
)

func newParseContext(expression string, namedVals []INamedValueInfo[INamedValue], fns []functions.IFnInfo[functions.IFn], allowUnknownKeyWords bool) *parseContext {
	result := parseContext{
		AllowUnknownKeywords: allowUnknownKeyWords,
		FnsInfo:              map[string]functions.IFnInfo[functions.IFn]{},
		NamedValsInfo:        map[string]INamedValueInfo[INamedValue]{},
	}
	if len(expression) > maxExpressionLength {
		panic(ErrorsMaxLengthExceeded)
	}

	for _, value := range namedVals {
		result.NamedValsInfo[value.GetName()] = value
	}

	for _, function := range fns {
		result.FnsInfo[function.GetName()] = function
	}
	result.Lexer = newLexer(expression)
	return &result
}

/*
Create Tree
	Create parseContext
		Create noop trace writer
		Create LexicalAnalyzer from Expression
		Add functions info to parseContext
		Add named values info to parseContext
		|-> createTree with parseContext

*/

func CreateTree(expression string, namedValues []INamedValueInfo[INamedValue],
	functions []functions.IFnInfo[functions.IFn]) (astRoot expression.IExpNode) {
	return createTree(newParseContext(expression, namedValues, functions, false))
}

func createTree(pCtx *parseContext) expression.IExpNode {
	for {
		token, haveToken := pCtx.Lexer.tryGetNextToken()
		pCtx.Token = token
		if !haveToken {
			break
		}
		if pCtx.Token.Kind() == lexicalTokenKindUnexpected {
			panic(fmt.Sprintf("unexpected token, rawValue: %s, kind: %s, expression: %s", pCtx.Token.RawValue(),
				pCtx.Token.Kind(),
				pCtx.Expression))
		}
		if pCtx.Token.IsOperator() {
			pushOperator(pCtx)
		} else {
			pushOperand(pCtx)
		}
		pCtx.LastToken = pCtx.Token
	}
	// no tokens
	if pCtx.LastToken == nil {
		return nil
	}
	// check unexpected end of expression
	if len(pCtx.Operators) > 0 {
		var unexpectedLastToken bool
		switch pCtx.LastToken.Kind() {
		// Legal
		case lexicalTokenKindEndGroup, lexicalTokenKindEndIndex, lexicalTokenKindEndParameters:
			break
			// Illegal
		case lexicalTokenKindFunction:
			unexpectedLastToken = true
		default:
			unexpectedLastToken = pCtx.LastToken.IsOperator()
		}
		if unexpectedLastToken || len(pCtx.Lexer.getUnclosedTokens()) > 0 {
			panic(fmt.Errorf("unexpected last token, rawValue: %s, kind: %s, expression: %s", pCtx.LastToken.RawValue(),
				pCtx.LastToken.Kind(),
				pCtx.Expression))
		}
	}
	for len(pCtx.Operators) > 0 {
		flushTopOperator(pCtx)
	}
	if len(pCtx.Operands) > 1 {
		panic("invalid number of operands")
	}
	root := pCtx.Operands[0].(expression.IExpNode)
	if err := checkMaxDepth(pCtx, root, 1); err != nil {
		panic(err)
	}
	return root
}

func pushOperator(pCtx *parseContext) {
	if pCtx.Token.Associativity() == associativityLTR {
		tk := pCtx.Token
		for len(pCtx.Operators) > 0 {
			topOp := pCtx.Operators[len(pCtx.Operators)-1]
			if topOp.Precedence() >= tk.Precedence() &&
				topOp.Kind() != lexicalTokenKindStartGroup &&
				topOp.Kind() != lexicalTokenKindStartIndex &&
				topOp.Kind() != lexicalTokenKindStartParameters &&
				topOp.Kind() != lexicalTokenKindSeparator {
				flushTopOperator(pCtx)
				continue
			}
			break
		}
	}
	pCtx.Operators = append(pCtx.Operators, pCtx.Token)
	// Process closing operators now, since context.LastToken is required
	// to accurately process TokenKind.lexicalTokenKindEndParameters
	switch pCtx.Token.Kind() {
	case lexicalTokenKindEndGroup, lexicalTokenKindEndIndex, lexicalTokenKindEndParameters:
		flushTopOperator(pCtx)
	}
}

func pushOperand(pCtx *parseContext) {
	switch pCtx.Token.Kind() {
	case lexicalTokenKindFunction:
		fn := pCtx.Token.RawValue()
		if fnInfo := tryGetFnInfo(pCtx, fn); fnInfo != nil {
			node := fnInfo.CreateNode().(functions.IFn).(expression.IExpNode)
			node.SetName(fn)
			pCtx.Operands = append(pCtx.Operands, node)
		} else {
			if pCtx.AllowUnknownKeywords {
				node := new(functions.NoOpFn)
				node.SetName(fn)
				pCtx.Operands = append(pCtx.Operands, node)
			} else {
				panic(fmt.Errorf("unrecognized function"))
			}
		}
	case lexicalTokenKindNamedValue:
		name := pCtx.Token.RawValue()
		if namedValInfo, exist := pCtx.NamedValsInfo[name]; exist {
			node := namedValInfo.CreateNode().(INamedValue).(expression.IExpNode)
			node.SetName(name)
			pCtx.Operands = append(pCtx.Operands, node)
		} else {
			if pCtx.AllowUnknownKeywords {
				node := new(noOpNamedValue)
				node.SetName(name)
				pCtx.Operands = append(pCtx.Operands, node)
			} else {
				panic(fmt.Errorf("unrecognized named value"))
			}
		}

	default:
		pCtx.Operands = append(pCtx.Operands, newNodeFromToken(pCtx.Token))
	}
}

func flushTopOperator(pCtx *parseContext) {
	switch pCtx.Operators[len(pCtx.Operators)-1].Kind() {
	case lexicalTokenKindEndIndex: // "]"
		flushTopEndIndex(pCtx)
		return
	case lexicalTokenKindEndGroup: // ")" logical grouping
		flushTopEndGroup(pCtx)
		return
	case lexicalTokenKindEndParameters: // ")" function call
		flushTopEndParameters(pCtx)
		return
	}
	// remove top operator
	tk := pCtx.Operators[len(pCtx.Operators)-1]
	pCtx.Operators = pCtx.Operators[:len(pCtx.Operators)-1]

	node := newNodeFromToken(tk).(expression.IContainer)
	operands := popOperands(pCtx, tk.OperandCount())
	for _, o := range operands {
		if _, isAnd := node.(*operators.And); isAnd {
			nestedAnd, isAnd := o.(*operators.And)
			if isAnd {
				for _, p := range nestedAnd.Parameters() {
					node.AddParameter(p)
				}
				continue
			}
		}
		if _, isOr := node.(*operators.Or); isOr {
			nestedOr, isOr := o.(*operators.Or)
			if isOr {
				for _, p := range nestedOr.Parameters() {
					node.AddParameter(p)
				}
				continue
			}
		}
		node.AddParameter(o)
	}
	// Push the node to the operand stack
	pCtx.Operands = append(pCtx.Operands, node)
}

func strictPopOnOperator(pCtx *parseContext, expectedKind lexicalTokenKind) (popped *lexicalToken) {
	top := pCtx.Operators[len(pCtx.Operators)-1]
	if top.Kind() != expectedKind {
		panic(fmt.Sprintf("expected operator %s to be of kind %s", expectedKind, top.Kind()))
	}
	pCtx.Operators = pCtx.Operators[:len(pCtx.Operators)-1]
	return top
}

// popOperands remove the number
func popOperands(pCtx *parseContext, count int) []expression.IExpNode {
	var result []expression.IExpNode
	for i := 0; i < count; i++ {
		result = append(result, pCtx.Operands[len(pCtx.Operands)-1])
		pCtx.Operands = pCtx.Operands[:len(pCtx.Operands)-1]
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
	// Pop end index
	strictPopOnOperator(pCtx, lexicalTokenKindEndIndex)
	// Pop start index
	tk := strictPopOnOperator(pCtx, lexicalTokenKindStartIndex)
	node := newNodeFromToken(tk).(expression.IContainer)

	ops := popOperands(pCtx, tk.OperandCount())
	for _, o := range ops {
		node.AddParameter(o)
	}
	pCtx.Operands = append(pCtx.Operands, node)
}

// flushTopEndGroup remove top logical group ")" from operator stack
func flushTopEndGroup(pCtx *parseContext) {
	strictPopOnOperator(pCtx, lexicalTokenKindEndGroup)
	strictPopOnOperator(pCtx, lexicalTokenKindStartGroup)
}

// flushTopEndParameters remove top end parameter ")" end of function call
func flushTopEndParameters(pCtx *parseContext) {
	tk := strictPopOnOperator(pCtx, lexicalTokenKindEndParameters)
	// Sanity check top tk is the current token
	if tk != pCtx.Token {
		panic(fmt.Errorf("expected popped token to be the current token"))
	}
	var fn functions.IFn
	// no parameter fn
	if pCtx.LastToken.Kind() == lexicalTokenKindStartParameters {
		// node already exist on operand stack
		fn = pCtx.Operands[len(pCtx.Operands)-1].(functions.IFn)
	} else {
		// parameter fn
		var seperatorCnt int
		for pCtx.Operators[len(pCtx.Operators)-1].Kind() == lexicalTokenKindSeparator {
			seperatorCnt++
			pCtx.Operators = pCtx.Operators[:len(pCtx.Operators)-1]
		}
		// eg: func(x,y,z) -> 2 separator -> 3 params
		fnOperands := popOperands(pCtx, seperatorCnt+1)
		// node already exist on operand stack
		fn = pCtx.Operands[len(pCtx.Operands)-1].(functions.IFn)
		// add operand to the fn node
		for _, operand := range fnOperands {
			fn.AddParameter(operand)
		}
	}
	strictPopOnOperator(pCtx, lexicalTokenKindStartParameters)
	fnInfo := tryGetFnInfo(pCtx, fn.GetName())
	if fnInfo != nil && pCtx.AllowUnknownKeywords {
	}
	if err := fnLimitCheck(fn, fnInfo); err != nil {
		panic(err)
	}
}

// update to return type maybe ?
func tryGetFnInfo(pCtx *parseContext, name string) (result functions.IFnInfo[functions.IFn]) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		result = nil
	// 	}
	// }()
	eFn, existEfn := newExpressionConstants().WellKnownFns[name]
	cFn, existCfn := pCtx.FnsInfo[name]
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

func fnLimitCheck(f functions.IFn, expected functions.IFnInfo[functions.IFn]) (err error) {
	if len(f.Parameters()) < expected.MinParameters() {
		err = ErrorsTooFewParameters
	}
	if len(f.Parameters()) > expected.MaxParameters() {
		err = ErrorsTooManyParameters
	}
	return err
}

func checkMaxDepth(pCtx *parseContext, node expression.IExpNode, depth int) (err error) {
	if depth > expression.MaxDepth {
		return ErrorsMaxDepthExceeded
	}
	if container, isContainer := node.(expression.IContainer); isContainer {
		for _, param := range container.Parameters() {
			_ = checkMaxDepth(pCtx, param, depth+1)
		}
	}
	return nil
}
