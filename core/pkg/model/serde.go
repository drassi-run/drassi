/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"slices"

	"github.com/go-viper/mapstructure/v2"
)

var hooks []mapstructure.DecodeHookFunc

func RegisterDecodeHook(fn mapstructure.DecodeHookFunc) {
	hooks = append(hooks, fn)
}

type DecodeOption func(config *mapstructure.DecoderConfig)

func Decode(source any, target any) error {
	opt := WithDecodeHook(true)
	return DecodeWithOptions(source, target, opt)
}

func WithDecodeHook(registered bool, h ...mapstructure.DecodeHookFunc) DecodeOption {
	if registered {
		h = slices.Concat(h, hooks)
		h = append(h, WeaklyString)
		h = append(h, decoderHook)
	}
	return func(config *mapstructure.DecoderConfig) {
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(h...)
	}
}

func DecodeWithOptions(source any, target any, opts ...DecodeOption) error {
	metadata := mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		Result:   target,
		TagName:  "actions",
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

var unmarshalers []*json.Unmarshalers

func RegisterUnmarshalInterface[T any](dis func(raw jsontext.Value) (T, error)) {
	u := unmarshalInterface(dis)
	unmarshalers = append(unmarshalers, u)
}

func unmarshalInterface[T any](dis func(raw jsontext.Value) (T, error)) *json.Unmarshalers {
	fn := func(d *jsontext.Decoder, val *T) error {
		var raw jsontext.Value
		if err := json.UnmarshalDecode(d, &raw); err != nil {
			return err
		}

		if t, err := dis(raw); err != nil {
			return err
		} else if err = json.Unmarshal(raw, t, d.Options()); err != nil {
			return err
		} else {
			*val = t
			return nil
		}
	}
	return json.UnmarshalFromFunc(fn)
}
