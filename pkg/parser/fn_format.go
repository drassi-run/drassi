package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrorsInvalidFormatArgIndex = errors.New("invalid format argument index")
)

type FormatFn struct {
	Fn
}

func (f *FormatFn) traceFullyRealized() bool {
	return false
}

func (f *FormatFn) setName(name string) {
	f.name = name
}

func (f *FormatFn) getName() string {
	return f.name
}

func (f *FormatFn) getContainer() iContainer {
	return f.container
}

func (f *FormatFn) setContainer(c iContainer) {
	f.container = c
}

func (f *FormatFn) evaluateCore(eCtx *EvaluationContext) any {
	format := evaluate(eCtx, f.Parameters()[0]).ConvertToString()
	var idx int
	result := newFormatResultBuilder(f, eCtx, len(format))
	for idx < len(format) {
		// Find first occurrence of opening brace and closing brace in sub-sequence from idx:len(format)-1
		lBraceIdx := strings.Index(format[idx:], "{") + idx
		rBraceIdx := strings.Index(format[idx:], "}") + idx
		// Left brace found
		if lBraceIdx >= idx && (rBraceIdx < idx || rBraceIdx > lBraceIdx) {
			// escaped left brace
			if at(format, lBraceIdx+1) == '{' {
				result.appendStatic(format[idx : lBraceIdx-idx+1])
				idx = lBraceIdx + 2
				continue
			}
			// Left brace, number, right brace
			if rBraceIdx > lBraceIdx+1 {
				if ok, argIdx := readArgIdx(format, lBraceIdx+1); ok {
					// Check parameter count
					if argIdx > len(f.Parameters())-2 {
						panic(ErrorsInvalidFormatArgIndex)
					}
					// Append the portion before the left brace
					result.appendStatic(format[idx:lBraceIdx])
					// Append the arg
					result.appendArgument(argIdx)
					idx = rBraceIdx + 1
				} else {
					panic(fmt.Errorf("invalid format string %s", format))
				}
			}
			continue
		}
		// Only right brace found
		if rBraceIdx >= idx {
			// escaped right brace
			if at(format, rBraceIdx+1) == '}' {
				result.appendStatic(format[idx : rBraceIdx+1])
				idx = rBraceIdx + 2
			} else {
				panic(fmt.Errorf("invalid format string %s", format))
			}
			continue
		}
		// Last segment
		result.appendStatic(format[idx:])
	}
	return result.String()
}

func readArgIdx(str string, start int) (ok bool, argIndex int) {
	var length int
	if unicode.IsDigit(at(str, start+length)) {
		length++
	}
	if length < 1 {
		return false, -1
	}
	argIndex, err := strconv.Atoi(str[start : start+length])
	if err != nil {
		return false, -1
	}
	return true, argIndex
}

func at(str string, i int) rune {
	s := []rune(str)
	if i < 0 || i >= len(str) {
		return rune(0)
	}
	return s[i]
}

type formatResultBuilder struct {
	node     *FormatFn
	ctx      *EvaluationContext
	segments []string
	cache    []*argValue
}

type argValue struct {
	stringResult string
}

func (a *argValue) StringResult() string {
	return a.stringResult
}

func newFormatResultBuilder(node *FormatFn, ctx *EvaluationContext, cacheSize int) *formatResultBuilder {
	return &formatResultBuilder{node: node, ctx: ctx, segments: []string{}, cache: make([]*argValue, cacheSize)}
}

func (f *formatResultBuilder) String() string {
	return strings.Join(f.segments, "")
}

func (f *formatResultBuilder) appendStatic(value string) {
	if len(value) > 0 {
		f.segments = append(f.segments, value)
	}
}

func (f *formatResultBuilder) appendArgument(argIdx int) {
	var result string
	argVal := f.cache[argIdx]
	if argVal == nil {
		// cache miss
		var evaluationResult = evaluate(f.ctx, f.node.Parameters()[argIdx+1])
		argVal = &argValue{
			stringResult: evaluationResult.ConvertToString(),
		}
		f.cache = append(f.cache, argVal)
	}
	result = argVal.StringResult()
	f.appendStatic(result)
}

func (f *FormatFn) convertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}

func (f *FormatFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}
