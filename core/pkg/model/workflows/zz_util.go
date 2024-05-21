package workflows

import (
	"reflect"
)

func valueOf(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
}
