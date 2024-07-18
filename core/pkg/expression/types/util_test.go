package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/model/workflows"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

type I interface{ foobar() }
type F float64
type L []string
type M map[int]int

type S struct {
	Boolean bool              `gha:"boolean"`
	Integer int64             `gha:"integer"`
	Float   float64           `gha:"float"`
	String  string            `gha:"string"`
	List    []string          `gha:"list"`
	Map     map[string]string `gha:"map"`

	Ignore   byte
	EmptyTag rune `gha:""`
}

// Pointer implement interface I
func (_ *F) foobar() {}
func (_ *L) foobar() {}
func (_ *M) foobar() {}
func (_ *S) foobar() {}

type FV float64
type LV []string
type MV map[int]int
type SV struct {
	Name string `gha:"name"`
}

// Value implement interface I
func (_ FV) foobar() {}
func (_ LV) foobar() {}
func (_ MV) foobar() {}
func (_ SV) foobar() {}

// New data type that have underlay type is a pointer
type FP *float64
type LP *[]string
type MP *map[int]int

func TestNativeToVal(t *testing.T) {
	t.Run("scala", testNativeToValScala)
	t.Run("collection", testNativeToValCollection)
	t.Run("dynamic", testNativeToValDynamic)
	t.Run("struct", testNativeToValStruct)
	t.Run("null", testNativeToValNull)
	t.Run("escalate", testNativeToValEscalate)
}

func testNativeToValScala(t *testing.T) {
	one := Float(1)
	f1 := 1.
	f2 := F(1)
	f3 := FV(1)
	f4 := FP(&f1)

	for _, v := range []any{
		f1, &f1, reflect.ValueOf(f1), reflect.ValueOf(&f1),
		f2, &f2, reflect.ValueOf(f2), reflect.ValueOf(&f2), valueOf[I](&f2),
		valueOf[I](f3),
		f4, &f4, reflect.ValueOf(f4), reflect.ValueOf(&f4),
	} {
		val := NativeToVal(v)
		assert.Equal(t, one, val, "convert from %#v (%[1]T)", v)
	}
}

func testNativeToValCollection(t *testing.T) {
	l1 := make([]string, 1)
	l2 := make(L, 1)
	l3 := make(LV, 1)
	x := make([]string, 2)
	l4 := LP(&x)

	for _, v := range []any{
		l1, &l1, reflect.ValueOf(l1), reflect.ValueOf(&l1),
		l2, &l2, reflect.ValueOf(l2), reflect.ValueOf(&l2), valueOf[I](&l2),
		valueOf[I](l3),
		l4, &l4, reflect.ValueOf(l4), reflect.ValueOf(&l4),
	} {
		val := NativeToVal(v)
		assert.Equal(t, ref.TypeList, val.Type(), "convert from %#v (%[1]T)", v)

		// test list is created from NewListGeneric
		getter := val.(*List).getter
		assert.Contains(t, functionName(getter), "NewListGeneric", "convert from %#v (%[1]T)", v)
	}
}

func testNativeToValDynamic(t *testing.T) {
	m1 := make(map[int]int) // map[int]int is not in predefined NativeToVal types
	m2 := make(M)
	m3 := make(MV)
	x := make(map[int]int, 2)
	m4 := MP(&x)

	for _, v := range []any{
		m1, &m1, reflect.ValueOf(m1), reflect.ValueOf(&m1),
		m2, &m2, reflect.ValueOf(m2), reflect.ValueOf(&m2), valueOf[I](&m2),
		valueOf[I](m3),
		m4, &m4, reflect.ValueOf(m4), reflect.ValueOf(&m4),
	} {
		val := NativeToVal(v)
		assert.Equal(t, ref.TypeMap, val.Type(), "convert from %#v (%[1]T)", v)

		// test map is created from NewMapDynamic
		acc := val.(*Map).mapAccessor
		_, ok := acc.(*dynamicMapAccessor)
		assert.True(t, ok, "DynamicMap, convert from %#v (%[1]T)", v)
	}
}

func testNativeToValStruct(t *testing.T) {
	s := S{}
	i1 := valueOf[I](&S{}) // Kind() = reflect.Interface, Interface() = *S{...}
	i2 := valueOf[I](SV{}) // Kind() = reflect.Interface, Interface() = SV{...}

	for _, p := range []workflows.KVPair[any, reflect.Kind]{
		{s, reflect.Struct},
		{&s, reflect.Pointer},
		{reflect.ValueOf(s), reflect.Struct},
		{reflect.ValueOf(&s), reflect.Pointer},
		{i1, reflect.Pointer},
		{i2, reflect.Struct},
	} {
		v, kind := p.Key, p.Value
		val := NativeToVal(v)
		assert.Equal(t, ref.TypeStruct, val.Type(), "convert from %#v (%[1]T)", v)
		assert.Equal(t, kind, reflect.ValueOf(val.Value()).Kind(), "convert from %#v (%[1]T)", v)
	}
}

func testNativeToValNull(t *testing.T) {
	for _, v := range []any{
		NULL, (I)(nil), reflect.ValueOf((I)(nil)),
		nil, reflect.ValueOf(nil), valueOf[I](nil),
		(*F)(nil), reflect.ValueOf((*F)(nil)), valueOf[I]((*F)(nil)),
		(*L)(nil), reflect.ValueOf((*L)(nil)), valueOf[I]((*L)(nil)),
		(*M)(nil), reflect.ValueOf((*M)(nil)), valueOf[I]((*M)(nil)),
		(*S)(nil), reflect.ValueOf((*S)(nil)), valueOf[I]((*S)(nil)),
		(LV)(nil), (*LV)(nil), reflect.ValueOf((LV)(nil)), reflect.ValueOf((*LV)(nil)), valueOf[I]((LV)(nil)), valueOf[I]((*LV)(nil)),
		(MV)(nil), (*MV)(nil), reflect.ValueOf((MV)(nil)), reflect.ValueOf((*MV)(nil)), valueOf[I]((MV)(nil)), valueOf[I]((*MV)(nil)),
	} {
		val := NativeToVal(v)
		assert.Equal(t, ref.TypeNull, val.Type(), "convert from %#v (%[1]T)", v)
	}
}

func testNativeToValEscalate(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		arr := [...]int{0, 1, 2, 3, 4, 5, 6}
		refArr := reflect.ValueOf(&arr).Elem()
		val := NativeToVal(refArr)
		assert.Equal(t, ref.TypeList, val.Type())
		// Value changed from array to slice
		assert.Equal(t, reflect.Slice, reflect.TypeOf(val.Value()).Kind())
	})

	t.Run("struct", func(t *testing.T) {
		str := S{}
		refS := reflect.ValueOf(&str).Elem()
		val := NativeToVal(refS)
		assert.Equal(t, ref.TypeStruct, val.Type())

		// Value changed from object to pointer
		refVal := reflect.ValueOf(val.Value())
		assert.Equal(t, reflect.Pointer, refVal.Kind())
		assert.Equal(t, reflect.Struct, refVal.Elem().Kind())
	})
}

func functionName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// reflect.ValueOf(x) cast x to any, so loose interface info
// This is a generic version of reflect.ValueOf to preserve this
func valueOf[V any](v V) reflect.Value {
	return reflect.ValueOf(&v).Elem()
}
