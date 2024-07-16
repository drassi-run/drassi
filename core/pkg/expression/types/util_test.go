package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

type I interface{ foobar() }
type F float64
type L []string
type M map[any]any

func TestNativeToVal(t *testing.T) {
	t.Run("scala", func(t *testing.T) {
		one := Float(1)
		x := 1.
		y := F(1)

		for _, v := range []any{
			x, &x, reflect.ValueOf(x), reflect.ValueOf(&x),
			y, &y, reflect.ValueOf(y), reflect.ValueOf(&y),
		} {
			z := NativeToVal(v)
			assert.Equal(t, one, z, "convert from %#v (%T)", v, v)
		}
	})

	t.Run("collection", func(t *testing.T) {
		l1 := make([]string, 2)
		l2 := make(L, 2)

		for _, v := range []any{
			l1, &l1, reflect.ValueOf(l1), reflect.ValueOf(&l1),
			l2, &l2, reflect.ValueOf(l2), reflect.ValueOf(&l2),
		} {
			z := NativeToVal(v)
			assert.EqualValues(t, ref.TypeList, z.Type(), "convert from %#v (%T)", v, v)

			// test list is created from NewListGeneric
			getter := z.(*List).getter
			assert.Contains(t, functionName(getter), "NewListGeneric", "convert from %#v (%T)", v, v)
		}
	})

	t.Run("dynamic", func(t *testing.T) {
		m1 := make(map[any]any, 2)
		m2 := make(M, 2)

		for _, v := range []any{
			m1, &m1, reflect.ValueOf(m1), reflect.ValueOf(&m1),
			m2, &m2, reflect.ValueOf(m2), reflect.ValueOf(&m2),
		} {
			z := NativeToVal(v)
			assert.EqualValues(t, ref.TypeMap, z.Type(), "convert from %#v (%T)", v, v)

			// test map is created from NewMapDynamic
			acc := z.(*Map).mapAccessor
			_, ok := acc.(*dynamicMapAccessor)
			assert.True(t, ok, "DynamicMap, convert from %#v (%T)", v, v)
		}
	})

	t.Run("null", func(t *testing.T) {
		for _, v := range []any{
			NULL, nil, reflect.ValueOf(nil),
			(*F)(nil), reflect.ValueOf((*F)(nil)),
			(I)(nil), reflect.ValueOf((I)(nil)),
		} {
			z := NativeToVal(v)
			assert.EqualValues(t, ref.TypeNull, z.Type(), "convert from %#v (%T)", v, v)
		}
	})
}

func functionName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
