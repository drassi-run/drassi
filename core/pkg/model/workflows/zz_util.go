package workflows

import (
	"fmt"
	"reflect"
)

func valueOf(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}

func castArray[E any](input []any) ([]E, error) {
	res := make([]E, 0, len(input))
	for _, v := range input {
		if e, ok := v.(E); ok {
			res = append(res, e)
			continue
		}
		return nil, fmt.Errorf("expected array of %s, got an element %#v type %T", reflect.TypeFor[E]().String(), v, v)
	}
	return res, nil
}

func castMap[K comparable, V any](input map[K]any) (map[K]V, error) {
	res := make(map[K]V, len(input))
	for k, v := range input {
		if e, ok := v.(V); ok {
			res[k] = e
			continue
		}
		return nil, fmt.Errorf("expected map of %s, got a value %#v type %T", reflect.TypeFor[V]().String(), v, v)
	}
	return res, nil
}
