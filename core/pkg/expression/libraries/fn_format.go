package libraries

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
)

const (
	passThrough = iota
	braceOpen
	braceClose
	placeHolder
)

var errInvalidFormat = errors.New("invalid format")

func Format(fmt ref.LazyVal, args ...ref.LazyVal) ref.Val {
	v := fmt()
	if v.Type() == ref.TypeInvalid {
		return v
	}
	format := stringify(v)
	argsCache := make([]*string, len(args))

	output := new(strings.Builder)
	output.Grow(len(format)) // best-effort to set output length

	state, idx := passThrough, -1
	for i, c := range format {
		//goland:noinspection GoDfaConstantCondition
		switch state {
		case passThrough: // normal buffer output
			switch c {
			case '{':
				state, idx = braceOpen, i+1
			case '}':
				state = braceClose
			default:
				output.WriteRune(c)
			}

		case braceOpen: // found '{'
			if c == '{' {
				// unescape "{{"
				output.WriteRune('{')
				state, idx = passThrough, -1
				break
			}

			state = placeHolder
			fallthrough

		case placeHolder:
			if c != '}' {
				// short-circuit to detect invalid format
				if !unicode.IsDigit(c) {
					return types.WrapError(errInvalidFormat)
				}
				break
			}
			// end placeholder, substitute it with replacement value
			index, err := strconv.Atoi(format[idx:i])
			if err != nil {
				return types.NewError("%w: %w", errInvalidFormat, err)
			}
			if index < 0 || index >= len(args) {
				return types.NewError("index out of range: %d in [0..%d)", index, len(args))
			}

			// compute args & cache result
			rep := argsCache[index]
			if rep == nil {
				v := args[index]()
				if v.Type() == ref.TypeInvalid {
					return v
				}
				s := stringify(v)
				rep = &s
				argsCache[index] = rep
			}
			output.WriteString(*rep)
			state, idx = passThrough, -1

		case braceClose: // found '}'
			if c == '}' {
				// unescape "}}"
				output.WriteRune('}')
				state, idx = passThrough, -1
				break
			}
			return types.WrapError(errInvalidFormat)
		}
	}

	switch state {
	case placeHolder, braceOpen:
		return types.NewError("unclosed brace at %d: %w", idx, errInvalidFormat)
	case braceClose:
		return types.NewError("closing brace without opening %d: %w", idx, errInvalidFormat)
	default:
		s := output.String()
		return types.String(s)
	}
}
