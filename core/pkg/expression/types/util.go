package types

import (
	"errors"
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
)

var errUncomparable = errors.New("uncomparable data types")
var errNaNCompare = errors.New("NaN values cannot be ordered")
var errInvalidType = errors.New("invalid data type")
var errUnsupportedType = errors.New("unsupported data type")

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
	case []byte:
		return String(v)

	// List generic
	case []bool:
		return NewListGeneric(v)
	case []string:
		return NewListGeneric(v)
	case []int:
		return NewListGeneric(v)
	case []int32:
		return NewListGeneric(v)
	case []int64:
		return NewListGeneric(v)
	case []float32:
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
	case map[string]int:
		return NewMapGeneric(v)
	case map[string]int32:
		return NewMapGeneric(v)
	case map[string]int64:
		return NewMapGeneric(v)
	case map[string]float32:
		return NewMapGeneric(v)
	case map[string]float64:
		return NewMapGeneric(v)
	case map[string]any:
		return NewMapGeneric(v)

	// Error
	case error:
		return WrapError(v)
	}

	// use reflect to also check primitive types (bool, int, float, string,...)
	// because above "switch rawVal.(type)" is not working on new type definition, e.g: `type A int`
	switch kind := refVal.Kind(); kind {
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

	case reflect.Map:
		// handle underlay datatype, e.g: `type M map[string]any`
		switch {
		case refVal.IsNil():
			return NULL
		case refVal.CanConvert(typeMapString):
			m := refVal.Convert(typeMapString).Interface().(map[string]string)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapInt):
			m := refVal.Convert(typeMapInt).Interface().(map[string]int)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapInt32):
			m := refVal.Convert(typeMapInt32).Interface().(map[string]int32)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapInt64):
			m := refVal.Convert(typeMapInt64).Interface().(map[string]int64)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapFloat32):
			m := refVal.Convert(typeMapFloat32).Interface().(map[string]float32)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapFloat64):
			m := refVal.Convert(typeMapFloat64).Interface().(map[string]float64)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapBool):
			m := refVal.Convert(typeMapBool).Interface().(map[string]bool)
			return NewMapGeneric(m)
		case refVal.CanConvert(typeMapAny):
			m := refVal.Convert(typeMapAny).Interface().(map[string]any)
			return NewMapGeneric(m)
		default:
			return NewMapDynamic(rawVal)
		}

	case reflect.Slice:
		// handle underlay datatype, e.g: `type L []string`
		switch {
		case refVal.IsNil():
			return NULL
		case refVal.CanConvert(typeListString):
			l := refVal.Convert(typeListString).Interface().([]string)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListInt):
			l := refVal.Convert(typeListInt).Interface().([]int)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListInt32):
			l := refVal.Convert(typeListInt32).Interface().([]int32)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListInt64):
			l := refVal.Convert(typeListInt64).Interface().([]int64)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListFloat32):
			l := refVal.Convert(typeListFloat32).Interface().([]float32)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListFloat64):
			l := refVal.Convert(typeListFloat64).Interface().([]float64)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListBool):
			l := refVal.Convert(typeListBool).Interface().([]bool)
			return NewListGeneric(l)
		case refVal.CanConvert(typeListAny):
			l := refVal.Convert(typeListAny).Interface().([]any)
			return NewListGeneric(l)
		default:
			return NewListDynamic(rawVal)
		}

	case reflect.Array:
		// storing slice is prefer to array
		// So, convert from array to slice when possible, e.g:
		// ```
		// arr := [...]int{0, 1, 2, 3, 4, 5, 6}
		// refArr := reflect.ValueOf(&arr).Elem()
		// ```
		if refVal.CanAddr() {
			refSlice := refVal.Slice(0, refVal.Len())
			return NativeToVal(refSlice)
		}
		return NewListDynamic(rawVal)

	case reflect.Struct:
		// storing pointer to struct is prefer to struct
		// So, get struct instance pointer when possible, e.g:
		// ```
		// s := MyStruct{}
		// refS := reflect.ValueOf(&s).Elem()
		// ```
		if refVal.CanAddr() {
			refStruct := refVal.Addr() // convert from Struct to *Struct
			rawVal = refStruct.Interface()
		}
		return NewStruct(rawVal)

	case reflect.Pointer, reflect.Interface:
		if refVal.IsNil() {
			return NULL
		}

		switch elem := refVal.Elem(); elem.Kind() {
		case reflect.Struct:
			return NewStruct(rawVal) // keep rawVal as a pointer
		default:
			return NativeToVal(elem)
		}

	default:
		return NewError("%w: %T", errUnsupportedType, rawVal)
	}
}

var (
	typeListString  = reflect.TypeFor[[]string]()
	typeListInt     = reflect.TypeFor[[]int]()
	typeListInt32   = reflect.TypeFor[[]int32]()
	typeListInt64   = reflect.TypeFor[[]int64]()
	typeListFloat32 = reflect.TypeFor[[]float32]()
	typeListFloat64 = reflect.TypeFor[[]float64]()
	typeListBool    = reflect.TypeFor[[]bool]()
	typeListAny     = reflect.TypeFor[[]any]()

	typeMapBool    = reflect.TypeFor[map[string]bool]()
	typeMapString  = reflect.TypeFor[map[string]string]()
	typeMapInt     = reflect.TypeFor[map[string]int]()
	typeMapInt32   = reflect.TypeFor[map[string]int32]()
	typeMapInt64   = reflect.TypeFor[map[string]int64]()
	typeMapFloat32 = reflect.TypeFor[map[string]float32]()
	typeMapFloat64 = reflect.TypeFor[map[string]float64]()
	typeMapAny     = reflect.TypeFor[map[string]any]()
)
