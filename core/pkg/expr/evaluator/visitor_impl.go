package evaluator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/mitchellh/mapstructure"

	"github.com/dungdm93/drassi/core/pkg/expr"
	"github.com/dungdm93/drassi/core/pkg/expr/ast"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/common"
	"github.com/dungdm93/drassi/core/pkg/model/contexts"
)

type visitorImpl struct {
	workingDir string
}

func (v visitorImpl) VisitAlwaysFn(c ast_ifaces.ExprNode) any {
	return true
}

func (v visitorImpl) VisitAnd(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	result := &result{}
	for _, param := range c.Parameters() {
		result = evaluate(eCtx, param)
		if result.isFalsy() {
			return result.Value()
		}
	}
	return result.Value()
}

func (v visitorImpl) VisitCancelledFn(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	tplCtx := eCtx.State().(*contexts.Expr)
	if tplCtx == nil {
		panic(ErrorExprContextNotFound)
	}
	execCtx := tplCtx.State
	if execCtx == nil {
		panic(ErrorExecutionContextNotFound)
	}
	return execCtx.Job.Status == contexts.ActionResultCancelled
}

func (v visitorImpl) VisitContainsFn(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.primitive() {
		lStr := l.string()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.primitive() {
			rStr := r.string()
			return containsIgnoreCase(lStr, rStr)
		}
	}
	isCol, col := l.isCollection()
	if isCol {
		if arr, isArr := col.(common.Array); isArr && len(arr) > 0 {
			r := evaluate(eCtx, c.Parameters()[1])
			for _, value := range arr {
				i := createIntermediateResult(eCtx, value)
				if r.equal(i) {
					return true
				}
			}
		}
	}
	return false
}

func (v visitorImpl) VisitContextValueNode(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	var target map[string]interface{}
	if err := mapstructure.Decode(*(eCtx.State().(*contexts.Expr).State), &target); err != nil {
		panic(err)
	}
	return target[c.GetName()]
}

func (v visitorImpl) VisitEndsWithFn(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.primitive() {
		lStr := l.string()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.primitive() {
			rStr := r.string()
			return endsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func (v visitorImpl) VisitEqual(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.equal(r)
}

// TODO: composite action
// See https://github.com/dungdm93/drasi/blob/bfc21ce03ad75998a64d4a4718c7d648fea24f2a/pkg/expression/evaluator/visitor.go#L99
func (v visitorImpl) VisitFailureFn(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	tplCtx := eCtx.State().(*contexts.Expr)
	if tplCtx == nil {
		panic(ErrorExprContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State
	return execCtx.Job.Status == contexts.ActionResultFailure
}

func (v visitorImpl) VisitFormatFn(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	format := evaluate(eCtx, c.Parameters()[0]).string()
	var idx int
	result := newFormatResultBuilder(c, eCtx, len(format))
	for idx < len(format) {
		// Find first occurrence of opening brace and closing brace in subsequence from idx:len(format)-1
		lBrace := strings.Index(format[idx:], "{") + idx
		rBrace := strings.Index(format[idx:], "}") + idx
		// Left brace found
		if lBrace >= idx && (rBrace < idx || rBrace > lBrace) {
			// escaped left brace
			if at(format, lBrace+1) == '{' {
				result.appendStatic(format[idx : lBrace-idx+1])
				idx = lBrace + 2
				continue
			}
			// Left brace, number, right brace
			ok1 := rBrace > lBrace+1
			ok2, argIdx := readArgIdx(format, lBrace+1)
			if !ok1 || !ok2 {
				panic(fmt.Errorf("invalid format string %s", format))
			}
			// Check parameter count
			if argIdx > len(c.Parameters())-2 {
				panic(ErrorInvalidFormatArgIndex)
			}
			// Append the portion before the left brace
			result.appendStatic(format[idx:lBrace])
			// Append the arg
			result.appendArgument(argIdx)
			idx = rBrace + 1
			continue
		}
		// Only right brace found
		if rBrace >= idx {
			// escaped right brace
			if at(format, rBrace+1) == '}' {
				result.appendStatic(format[idx : rBrace+1])
				idx = rBrace + 2
			} else {
				panic(fmt.Errorf("invalid format string %s", format))
			}
			continue
		}
		// Last segment
		result.appendStatic(format[idx:])
		break
	}
	return result.String()
}

func (v visitorImpl) VisitFromJsonFn(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	val := evaluate(eCtx, c.Parameters()[0])
	if val.kind != expr.String {
		panic(fmt.Errorf("Cannot parse non-string type %v as JSON", val.kind))
	}
	var data any
	err := json.Unmarshal([]byte(val.string()), &data)
	if err != nil {
		panic(fmt.Errorf("Invalid json, err: %+v", err))
	}
	return data
}

func (v visitorImpl) VisitGreaterThan(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.greaterThan(r)
}

func (v visitorImpl) VisitGreaterThanOrEqual(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.greaterThanOrEqual(r)
}

func (v visitorImpl) VisitIndex(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	isCol, col := l.isCollection()
	if !isCol {
		_, isW := c.Parameters()[1].(*ast.WildCard)
		if isW {
			return newFilteredArray()
		}
		return nil
	}
	fa, isFilteredArray := col.(filteredArray)
	if isFilteredArray {
		return handleFilteredArray(eCtx, fa, c)
	}
	obj, isObj := col.(common.Obj)
	if isObj {
		return handleObject(eCtx, obj, c)
	}
	arr, isArr := col.(common.Array)
	if isArr {
		return handleArray(eCtx, arr, c)
	}
	return nil
}

func (v visitorImpl) VisitJoinFn(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	items := evaluate(eCtx, c.Parameters()[0])
	isCol, col := items.isCollection()
	if isCol {
		arr, isArr := col.(common.Array)
		if isArr && len(arr) > 0 {
			var result strings.Builder
			item := arr[0]
			itemResult := createIntermediateResult(eCtx, item)
			itemStr := itemResult.string()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if len(arr) > 1 {
				separator := ","
				if len(c.Parameters()) > 1 {
					separatorResult := evaluate(eCtx, c.Parameters()[1])
					if separatorResult.primitive() {
						separator = separatorResult.string()
					}
				}
				for i :=1; i<  len(arr); i++ {
					result.WriteString(separator)
					nextItem := arr[i]
					nextItemResult := createIntermediateResult(eCtx, nextItem)
					nextItemStr := nextItemResult.string()
					result.WriteString(nextItemStr)
				}
			}
			return result.String()
		}
	}
	if items.primitive() {
		return items.string()
	}
	return ""
}

func (v visitorImpl) VisitLessThan(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	left := evaluate(eCtx, c.Parameters()[0])
	right := evaluate(eCtx, c.Parameters()[1])
	return left.lessThan(right)
}

func (v visitorImpl) VisitLessThanOrEqual(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	left := evaluate(eCtx, c.Parameters()[0])
	right := evaluate(eCtx, c.Parameters()[1])
	return left.lessThanOrEqual(right)
}

func (v visitorImpl) VisitLiteral(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	return c.Value()
}

func (v visitorImpl) VisitNoopFn(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	return nil
}

func (v visitorImpl) VisitNoopNamedValue(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	return nil
}

func (v visitorImpl) VisitNot(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	return evaluate(eCtx, c.Parameters()[0]).isFalsy()
}

func (v visitorImpl) VisitNotEqual(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.notEqual(r)
}

func (v visitorImpl) VisitOr(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	var result *result
	for _, p := range c.Parameters() {
		result = evaluate(eCtx, p)
		if result.isTruthy() {
			break
		}
	}
	if result == nil {
		return nil
	}
	return result.Value()
}

func (v visitorImpl) VisitStartsWithFn(eCtx ast_ifaces.Context, c ast_ifaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.primitive() {
		lStr := l.string()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.primitive() {
			rStr := r.string()
			return startsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

// TODO: composite action
// See https://github.com/dungdm93/drasi/blob/bfc21ce03ad75998a64d4a4718c7d648fea24f2a/pkg/expression/evaluator/visitor.go#L320
func (v visitorImpl) VisitSuccessFn(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	tplCtx := eCtx.State().(*contexts.Expr)
	if tplCtx == nil {
		panic(ErrorExprContextNotFound)
	}
	ctx := tplCtx.State
	return ctx.Job.Status == contexts.ActionResultSuccess
}

func (v visitorImpl) VisitWildCard(eCtx ast_ifaces.Context, c ast_ifaces.ExprNode) any {
	return common.Wildcard
}

// See https://github.com/nektos/act/blob/8acde99bfa9cd91ad5561a21253965429a6a101f/pkg/exprparser/functions.go#L183
// and https://github.com/nektos/act/blob/8acde99bfa9cd91ad5561a21253965429a6a101f/pkg/runner/expression.go#L160
func (v visitorImpl) VisitHashfileFn(eCtx ast_ifaces.Context, n ast_ifaces.Container) any {
	var ps []gitignore.Pattern
	const cwdPrefix = "." + string(filepath.Separator)
	const excludeCwdPrefix = "!" + cwdPrefix
	for _, p := range n.Parameters() {
		path := evaluate(eCtx, p)
		if path.Kind() == expr.String {
			cleanPath := path.string()
			if strings.HasPrefix(cleanPath, cwdPrefix) {
				cleanPath = cleanPath[len(cwdPrefix):]
			} else if strings.HasPrefix(cleanPath, excludeCwdPrefix) {
				cleanPath = "!" + cleanPath[len(excludeCwdPrefix):]
			}
			ps = append(ps, gitignore.ParsePattern(cleanPath, nil))
		} else {
			panic(fmt.Errorf("Non-string path passed to hashFiles"))
		}
	}

	matcher := gitignore.NewMatcher(ps)

	var files []string

	if err := filepath.Walk(v.workingDir, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		sansPrefix := strings.TrimPrefix(path, v.workingDir+string(filepath.Separator))
		parts := strings.Split(sansPrefix, string(filepath.Separator))
		if fi.IsDir() || !matcher.Match(parts, fi.IsDir()) {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		panic(fmt.Errorf("Unable to filepath.Walk: %v", err))
	}

	if len(files) == 0 {
		return ""
	}

	hasher := sha256.New()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			panic(fmt.Errorf("Unable to os.Open: %v", err))
		}

		if _, err := io.Copy(hasher, f); err != nil {
			panic(fmt.Errorf("Unable to io.Copy: %v", err))
		}

		if err := f.Close(); err != nil {
			panic(fmt.Errorf("Unable to Close file: %v", err))
		}
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func containsIgnoreCase(s string, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func endsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

func startsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(suffix))
}
