/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"bytes"
	"encoding/json/v2"
	"io"

	"github.com/pelletier/go-toml/v2"
)

// Marshal encodes v into TOML format with the marshaler interface enabled.
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := MarshalWrite(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MarshalWrite(w io.Writer, v any) error {
	enc := toml.NewEncoder(w).
		SetIndentTables(true).
		EnableMarshalerInterface()
	return enc.Encode(v)
}

// Unmarshal decodes TOML data into v with the unmarshaler interface enabled.
func Unmarshal(data []byte, v any) error {
	var r = bytes.NewReader(data)
	return UnmarshalRead(r, v)
}

func UnmarshalRead(r io.Reader, v any) error {
	dec := toml.NewDecoder(r).
		EnableUnmarshalerInterface()
	return dec.Decode(v)
}

func (s *Sandboxer) MarshalTOML() ([]byte, error) {
	m := make(map[string]any)
	if len(s.Config) > 0 {
		if err := json.Unmarshal(s.Config, &m); err != nil {
			return nil, err
		}
	}
	if s.Provider != "" {
		m["provider"] = s.Provider
	}
	return toml.Marshal(m)
}

func (s *Sandboxer) UnmarshalTOML(data []byte) error {
	var m map[string]any
	if err := toml.Unmarshal(data, &m); err != nil {
		return err
	}
	if v, ok := m["provider"].(string); ok {
		s.Provider = v
		delete(m, "provider")
	}
	if len(m) > 0 {
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		s.Config = b
	} else {
		s.Config = nil
	}
	return nil
}
