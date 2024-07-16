package types

import (
	"errors"
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
)

var errUncomparable = errors.New("uncomparable data types")
var errNaNCompare = errors.New("NaN values cannot be ordered")

func NativeToVal(val any) ref.Val {
	if val == nil {
		return NULL
	}
	if v, ok := val.(ref.Val); ok {
		return v
	}

	var rawVal = val
	var refVal reflect.Value
	if v, ok := val.(reflect.Value); ok {
		if !v.IsValid() {
			return NULL
		}
		rawVal = v.Interface()
		if rawVal == nil {
			return NULL
		}
		refVal = v
	} else {
		refVal = reflect.ValueOf(val)
	}
	if refVal.Kind() == reflect.Ptr {
		if refVal.IsNil() {
			return NULL
		}
		refVal = refVal.Elem()
	}

	// rawVal & refVal both non-null & is not ref.Val
	switch v := rawVal.(type) {
	case bool:
		return Boolean(v)
	case int:
		return Integer(v)
	case int8:
		return Integer(v)
	case int16:
		return Integer(v)
	case int32:
		return Integer(v)
	case int64:
		return Integer(v)
	case uint:
		return Integer(v)
	case uint8:
		return Integer(v)
	case uint16:
		return Integer(v)
	case uint32:
		return Integer(v)
	case uint64:
		return Integer(v)
	case float32:
		return Float(v)
	case float64:
		return Float(v)
	case string:
		return String(v)

	case *bool:
		return Boolean(*v)
	case *int:
		return Integer(*v)
	case *int8:
		return Integer(*v)
	case *int16:
		return Integer(*v)
	case *int32:
		return Integer(*v)
	case *int64:
		return Integer(*v)
	case *uint:
		return Integer(*v)
	case *uint8:
		return Integer(*v)
	case *uint16:
		return Integer(*v)
	case *uint32:
		return Integer(*v)
	case *uint64:
		return Integer(*v)
	case *float32:
		return Float(*v)
	case *float64:
		return Float(*v)
	case *string:
		return String(*v)

	case []byte:
		return String(v)

	// List generic
	case []bool:
		return NewListGeneric(v)
	case []string:
		return NewListGeneric(v)
	case []int64:
		return NewListGeneric(v)
	case []float64:
		return NewListGeneric(v)
	case []any:
		return NewListGeneric(v)

	// Map generic
	case map[string]bool:
		return NewMapGeneric(v)
	case map[string]string:
		return NewMapGeneric(v)
	case map[string]int64:
		return NewMapGeneric(v)
	case map[string]float64:
		return NewMapGeneric(v)
	case map[string]any:
		return NewMapGeneric(v)
	}

	// use reflect to also check primitive types (bool, int, float, string,...)
	// because above "switch rawVal.(type)" is not working on new type definition, e.g: `type A int`
	switch refVal.Kind() {
	case reflect.Bool:
		b := refVal.Bool()
		return Boolean(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := refVal.Int()
		return Integer(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i := refVal.Uint()
		return Integer(i)
	case reflect.Float32, reflect.Float64:
		f := refVal.Float()
		return Float(f)
	case reflect.String:
		s := refVal.String()
		return String(s)
	case reflect.Struct:
		return nil // TODO

	case reflect.Map:
		switch {
		case refVal.CanConvert(typeMapString):
			m := refVal.Convert(typeMapString).Interface().(map[string]string)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapInteger):
			m := refVal.Convert(typeMapInteger).Interface().(map[string]int64)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapFloat):
			m := refVal.Convert(typeMapFloat).Interface().(map[string]float64)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapBool):
			m := refVal.Convert(typeMapBool).Interface().(map[string]bool)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapAny):
			m := refVal.Convert(typeMapAny).Interface().(map[string]any)
			return NewMapGeneric(m)
		default:
			return NewMapDynamic(refVal.Interface())
		}

	case reflect.Slice, reflect.Array:
		switch {
		case refVal.CanConvert(typeListString):
			l := refVal.Convert(typeListString).Interface().([]string)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListInteger):
			l := refVal.Convert(typeListInteger).Interface().([]int64)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListFloat):
			l := refVal.Convert(typeListFloat).Interface().([]float64)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListBool):
			l := refVal.Convert(typeListBool).Interface().([]bool)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListAny):
			l := refVal.Convert(typeListAny).Interface().([]any)
			return NewListGeneric(l)
		default:
			return NewListDynamic(refVal.Interface())
		}

	default:
		return nil // Not supported data type
	}
}

var (
	typeListString  = reflect.TypeFor[[]string]()
	typeListInteger = reflect.TypeFor[[]int64]()
	typeListFloat   = reflect.TypeFor[[]float64]()
	typeListBool    = reflect.TypeFor[[]bool]()
	typeListAny     = reflect.TypeFor[[]any]()

	typeMapBool    = reflect.TypeFor[map[string]bool]()
	typeMapString  = reflect.TypeFor[map[string]string]()
	typeMapInteger = reflect.TypeFor[map[string]int64]()
	typeMapFloat   = reflect.TypeFor[map[string]float64]()
	typeMapAny     = reflect.TypeFor[map[string]any]()
)
