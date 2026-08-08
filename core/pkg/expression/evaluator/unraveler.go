/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package evaluator

import (
	"errors"
	"fmt"
	"iter"
	"regexp"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	"drassi.run/core/pkg/model/workflows"
)

type unraveler struct {
	env expression.Env
}

func (u *unraveler) UnravelLiteral(val any) (ref.Val, error) {
	return types.NativeToVal(val), nil
}

func (u *unraveler) UnravelExpression(expr string, pure bool) (ref.Val, error) {
	node, err := u.env.Parse(expr, pure)
	if err != nil {
		return nil, err
	}

	prog, err := u.env.Bind(node)
	if err != nil {
		return nil, err
	}

	return prog.Execute()
}

func (u *unraveler) UnravelSequence(seq []workflows.Token) (_ []ref.Val, err error) {
	res := make([]ref.Val, 0, len(seq))

	errorHandler := func(e error) bool {
		err = e
		return false
	}

	for v := range u.sequenceIterator(seq, errorHandler) {
		res = append(res, v)
	}

	if err != nil {
		return
	}
	return res, nil
}

func (u *unraveler) sequenceIterator(
	seq []workflows.Token,
	onError func(err error) bool,
) iter.Seq[ref.Val] {
	return func(yield func(ref.Val) bool) {
		for _, token := range seq {
			val, err := token.Unravel(u)
			if err != nil {
				if onError(err) {
					continue
				}
				return
			}

			if _, ok := workflows.Expression(token); !ok {
				if !yield(val) {
					return
				}
				continue
			}

			if val.Type() != ref.TypeList {
				if !yield(val) {
					return
				}
				continue
			}

			// expression result is a list, perform merging
			list := val.(traits.Iterable)
			for _, e := range list.Items() {
				if !yield(e) {
					return
				}
			}
		}
	}
}

func (u *unraveler) UnravelMapping(pairs [][2]workflows.Token) (_ map[string]ref.Val, err error) {
	errs := make([]error, 0)
	res := make(map[string]ref.Val, len(pairs))

	errorHandler := func(e error) bool {
		errs = append(errs, e)
		return true
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
	onError func(error) bool,
) iter.Seq2[string, ref.Val] {
	return func(yield func(string, ref.Val) bool) {
		for k, v := range tuple2(pairs) {
			val, err := v.Unravel(u)
			if err != nil {
				if !onError(err) {
					return
				}
				continue
			}

			if isInsertMode(k) {
				if typ := val.Type(); typ != ref.TypeMap && typ != ref.TypeStruct {
					err = fmt.Errorf("can't merge map with %q", typ)
					if !onError(err) {
						return
					}
					continue
				}

				dict := val.(traits.Iterable)
				for kItem, vItem := range dict.Items() {
					if s, ok := kItem.(traits.Stringable); !ok {
						err = fmt.Errorf("key %v can't convert to string", kItem)
						if !onError(err) {
							return
						}
						continue
					} else if !yield(s.ToString(), vItem) {
						return
					}
				}
				continue
			}

			if key, err := k.Unravel(u); err != nil {
				if !onError(err) {
					return
				}
				continue
			} else if s, ok := key.(traits.Stringable); !ok {
				err = fmt.Errorf("key %v can't convert to string", key)
				if !onError(err) {
					return
				}
				continue
			} else if !yield(s.ToString(), val) {
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
