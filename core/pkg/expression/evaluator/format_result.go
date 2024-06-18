package evaluator

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type formatResultBuilder struct {
	node     ast_ifaces.Container
	ctx      ast_ifaces.Context
	segments []string
	cache    []*argValue
}

type argValue struct {
	stringResult string
}

func (a *argValue) StringResult() string {
	return a.stringResult
}

func newFormatResultBuilder(node ast_ifaces.Container, ctx ast_ifaces.Context, cacheSize int) *formatResultBuilder {
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
			stringResult: evaluationResult.string(),
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

func readFormatSpecifiers(str string, start int) (ok bool, result string, rBrace int) {
	c := at(str, start)
	if c == '}' {
		return true, result, start
	}
	if c != ':' {
		return false, result, rBrace
	}
	var specifiers strings.Builder
	idx := start + 1
	for {
		if idx >= len(str) {
			return false, result, rBrace
		}
		c = rune(str[idx])
		// Not right-brace
		if c != '}' {
			specifiers.WriteRune(c)
			idx++
			continue
		}
		// Escaped right-brace
		if at(str, idx+1) == '}' {
			specifiers.WriteRune('}')
			idx = idx + 2
			continue
		}
		return true, specifiers.String(), idx
	}
}

func at(str string, i int) rune {
	s := []rune(str)
	if i < 0 || i >= len(str) {
		return rune(0)
	}
	return s[i]
}
