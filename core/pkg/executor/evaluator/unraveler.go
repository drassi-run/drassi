package evaluator

import (
	"errors"
	"fmt"
	"iter"
	"reflect"
	"regexp"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
)

type unraveler struct {
	env expression.Env
}

func (u *unraveler) UnravelLiteral(val any) (any, error) {
	return val, nil
}

func (u *unraveler) UnravelExpression(expr string, pure bool) (any, error) {
	node, err := u.env.Parse(expr, pure)
	if err != nil {
		return nil, err
	}

	prog, err := u.env.Bind(node)
	if err != nil {
		return nil, err
	}

	return u.env.Execute(prog)
}

func (u *unraveler) UnravelSequence(seq []workflows.Token) (any, error) {
	errs := make([]error, 0)
	res := make([]any, 0, len(seq))

	errorHandler := func(e error) {
		errs = append(errs, e)
	}

	for v := range u.sequenceIterator(seq, errorHandler) {
		res = append(res, v)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return res, nil
}

func (u *unraveler) sequenceIterator(
	seq []workflows.Token,
	onError func(err error),
) iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, token := range seq {
			val, err := token.Unravel(u)
			if err != nil {
				onError(err)
				continue
			}

			if _, ok := workflows.Expression(token); !ok {
				if !yield(val) {
					return
				}
				continue
			}

			v := reflect.ValueOf(val)
			if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
				if !yield(val) {
					return
				}
				continue
			}

			// expression result is a list, perform merging
			for i := 0; i < v.Len(); i++ {
				e := v.Index(i).Interface()
				if !yield(e) {
					return
				}
			}
		}
	}
}

func (u *unraveler) UnravelMapping(pairs [][2]workflows.Token) (any, error) {
	errs := make([]error, 0)
	res := make(map[string]any, len(pairs))

	errorHandler := func(e error) {
		errs = append(errs, e)
	}

	for k, v := range u.mappingIterator(pairs, errorHandler) {
		if _, found := res[k]; found {
			errs = append(errs, fmt.Errorf("key duplicate: %s", k))
		}
		res[k] = v
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return res, nil
}

func (u *unraveler) mappingIterator(
	pairs [][2]workflows.Token,
	onError func(error),
) iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range tuple2(pairs) {
			insertMode := isInsertMode(k)
			val, err := v.Unravel(u)
			if err != nil {
				onError(err)
				continue
			}

			if insertMode {
				m, ok := model.ObjectStringify(val)
				if !ok {
					onError(fmt.Errorf("can't merge map with %T", val))
					continue
				}

				for kItem, vItem := range m {
					if !yield(kItem, vItem) {
						return
					}
				}
				continue
			}

			if key, err := k.Unravel(u); err != nil {
				onError(err)
			} else if s, ok := model.Stringify(key); !ok {
				onError(fmt.Errorf("key not a string: %T", key))
			} else if !yield(s, val) {
				return
			}
		}
	}
}

var insertModeRegex = regexp.MustCompile(`^\${{\s*insert\s*}}$`)

func isInsertMode(token workflows.Token) bool {
	expr, ok := workflows.Expression(token)
	return ok && insertModeRegex.MatchString(expr)
}

func tuple2[E any](pairs [][2]E) iter.Seq2[E, E] {
	return func(yield func(E, E) bool) {
		for _, pair := range pairs {
			k, v := pair[0], pair[1]
			if !yield(k, v) {
				return
			}
		}
	}
}
