package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"reflect"
	"testing"
)

func TestEquals(t *testing.T) {
	for _, l := range values {
		lhs := types.NativeToVal(l.value)
		for _, r := range values {
			rhs := types.NativeToVal(r.value)

			expected := isEquals(l, r)
			actual := Equals(lhs, rhs)
			verify(t, expected, actual, "%v == %v (%[1]T, %[2]T)", l.value, r.value)
		}
	}
}

func TestNotEquals(t *testing.T) {
	for _, l := range values {
		lhs := types.NativeToVal(l.value)
		for _, r := range values {
			rhs := types.NativeToVal(r.value)

			expected := !isEquals(l, r)
			actual := NotEquals(lhs, rhs)
			verify(t, expected, actual, "%v == %v (%[1]T, %[2]T)", l.value, r.value)
		}
	}
}

func isEquals(a, b weakTypeConversion) bool {
	if a.value == nil || b.value == nil {
		return a.numberify == b.numberify
	}

	va := reflect.ValueOf(a.value)
	vb := reflect.ValueOf(b.value)
	if va.Type() != vb.Type() {
		return a.numberify == b.numberify
	}
	switch va.Kind() {
	case reflect.Map, reflect.Slice, reflect.Pointer:
		return va.UnsafePointer() == vb.UnsafePointer()
	default:
		return a.value == b.value
	}
}
