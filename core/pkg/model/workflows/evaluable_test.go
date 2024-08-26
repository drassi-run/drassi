package workflows

import (
	"drassi.run/core/pkg/model"
	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/maps"
	"gotest.tools/v3/assert"
	"reflect"
	"slices"
	"testing"
)

func clone[T any](i T) T {
	if o, err := copystructure.Copy(i); err != nil {
		return i
	} else {
		return o.(T)
	}
}

func mapValues[E any](m map[string]E) []E {
	keys := maps.Keys(m)
	slices.Sort(keys)

	l := make([]E, len(keys))
	for i, k := range keys {
		l[i] = m[k]
	}
	return l
}

func mapTokenPairs[K comparable](m map[K]Token, keyConv func(K) Token) [][2]Token {
	r := make([][2]Token, 0)
	for k, v := range m {
		key := keyConv(k)
		r = append(r, [2]Token{key, v})
	}
	return r
}

func comparerForLiteralToken(opts ...cmp.Option) cmp.Option {
	return cmp.Comparer(func(x, y literalToken) bool {
		return cmp.Equal(x.value, y.value, opts...)
	})
}

func TestDecodeEvaluable(t *testing.T) {
	t.Run("kind", func(tt *testing.T) {
		type testEvaluable struct {
			LitBool   Evaluable[bool]           `actions:"litBool"`
			LitInt    Evaluable[int64]          `actions:"litInt"`
			LitFloat  Evaluable[float64]        `actions:"litFloat"`
			LitString Evaluable[string]         `actions:"litString"`
			Expr      Evaluable[string]         `actions:"expr"`
			Seq       Evaluable[[]any]          `actions:"seq"`
			Dict      Evaluable[map[string]any] `actions:"dict"`
		}

		scala := map[string]any{
			"litBool":   true,
			"litInt":    int64(123),
			"litFloat":  float64(1.23),
			"litString": "hello world",
			"expr":      "${{ foo.bar }}",
		}
		input := clone(scala)
		input["seq"] = mapValues(scala)
		input["dict"] = clone(scala)

		scalaExpected := map[string]Token{
			"litBool":   NewLiteralToken(true),
			"litInt":    NewLiteralToken(int64(123)),
			"litFloat":  NewLiteralToken(float64(1.23)),
			"litString": NewLiteralToken("hello world"),
			"expr":      NewExpressionToken("${{ foo.bar }}"),
		}
		expected := &testEvaluable{
			LitBool:   NewLiteralToken(true),
			LitInt:    NewLiteralToken(int64(123)),
			LitFloat:  NewLiteralToken(float64(1.23)),
			LitString: NewLiteralToken("hello world"),
			Expr:      NewExpressionToken("${{ foo.bar }}"),
			Seq:       NewSequenceToken(mapValues(scalaExpected)),
			Dict: NewMappingToken(mapTokenPairs(scalaExpected, func(s string) Token {
				return NewLiteralToken(s)
			})),
		}

		actual := new(testEvaluable)
		err := model.Decode(input, actual)
		assert.NilError(tt, err)

		opts := comparerForLiteralToken()
		assert.DeepEqual(tt, expected.Seq, actual.Seq, opts)
	})

	t.Run("invalid-reflection", func(tt *testing.T) {
		opt := func(config *mapstructure.DecoderConfig) {
			fault := func(from reflect.Value, to reflect.Value) (any, error) {
				if !to.Type().Implements(typeToken) {
					return valueOf(from), nil
				}
				return nil, nil // fault injection
			}
			// DecodeTokenHook MUST be after fault
			config.DecodeHook = mapstructure.ComposeDecodeHookFunc(fault, DecodeTokenHook)
		}

		token := new(literalToken)
		err := model.DecodeWithOptions("foobar", &token, opt)
		assert.NilError(tt, err)
	})
}
