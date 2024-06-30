package model

import (
	"reflect"
	"slices"

	"github.com/mitchellh/mapstructure"
)

var hooks []mapstructure.DecodeHookFunc

func RegisterDecodeHook(fn mapstructure.DecodeHookFunc) {
	hooks = append(hooks, fn)
}

// comparable to yaml.Unmarshaler, decoder allow a type to define its own custom logic to convert value
// see https://github.com/mitchellh/mapstructure/pull/294
type decoder interface {
	DecodeMapstructure(any) (any, error)
}

// see https://github.com/mitchellh/mapstructure/issues/115#issuecomment-735287466
// adapted to support types derived from built-in types, as DecodeMapstructure would not be able to mutate internal
// value, so need to invoke DecodeMapstructure defined by pointer to type
func decoderHook(from reflect.Value, to reflect.Value) (any, error) {
	// If the destination implements the decoder interface
	u, ok := to.Interface().(decoder)
	if !ok {
		// for non-struct types we need to invoke func (*type) DecodeMapstructure()
		if to.CanAddr() {
			pto := to.Addr()
			u, ok = pto.Interface().(decoder)
		}
		if !ok {
			return from.Interface(), nil
		}
	}
	// If it is nil and a pointer, create and assign the target value first
	if to.Type().Kind() == reflect.Ptr && to.IsNil() {
		to.Set(reflect.New(to.Type().Elem()))
		u = to.Interface().(decoder)
	}
	// Call the custom DecodeMapstructure method
	if d, err := u.DecodeMapstructure(from.Interface()); err != nil {
		return nil, err
	} else if d != nil {
		return d, nil
	} else {
		// d == nil: all input already processed
		return u, nil
	}
}

func init() {
	RegisterDecodeHook(decoderHook)
}

type DecodeOption func(config *mapstructure.DecoderConfig)

func Decode(source any, target any) error {
	opt := WithDecodeHook(true)
	return DecodeWithOptions(source, target, opt)
}

func WithDecodeHook(registered bool, h ...mapstructure.DecodeHookFunc) DecodeOption {
	if registered {
		h = slices.Concat(h, hooks)
	}
	return func(config *mapstructure.DecoderConfig) {
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(h...)
	}
}

func DecodeWithOptions(source any, target any, opts ...DecodeOption) error {
	metadata := mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		Result:   target,
		TagName:  "mapstructure",
		Metadata: &metadata,
	}
	for _, o := range opts {
		o(config)
	}

	d, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return d.Decode(source)
}
