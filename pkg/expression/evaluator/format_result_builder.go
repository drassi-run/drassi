package evaluator

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type formatResultBuilder struct {
	node     interfaces.IContainer
	ctx      interfaces.IEvaluationContext
	segments []string
	cache    []*argValue
}

type argValue struct {
	stringResult string
}

func (a *argValue) StringResult() string {
	return a.stringResult
}

func newFormatResultBuilder(node interfaces.IContainer, ctx interfaces.IEvaluationContext, cacheSize int) *formatResultBuilder {
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
		var evaluationResult = EvaluateWithContext(f.ctx, f.node.Parameters()[argIdx+1])
		argVal = &argValue{
			stringResult: evaluationResult.ConvertToString(),
		}
		f.cache = append(f.cache, argVal)
	}
	result = argVal.StringResult()
	f.appendStatic(result)
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
