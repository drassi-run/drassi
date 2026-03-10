/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"

	"drassi.run/core/pkg/model"
)

func init() {
	model.RegisterUnmarshalInterface(discriminateJob)
}

func discriminateJob(raw jsontext.Value) (Job, error) {
	var dis struct {
		RunsOn jsontext.Value `json:"runs-on,omitempty"`
		Uses   string         `json:"uses,omitempty"`
	}
	if err := json.Unmarshal(raw, &dis); err != nil {
		return nil, err
	}

	hasRunsOn := len(dis.RunsOn) > 0 && dis.RunsOn.Kind() != jsontext.KindNull
	hasUses := dis.Uses != ""

	switch {
	case hasRunsOn && hasUses:
		return nil, fmt.Errorf("job MUST be contains either `runs-on` or `uses`")
	case hasUses:
		return new(ReusableWorkflowCallJob), nil
	case hasRunsOn:
		return new(NormalJob), nil
	default:
		// both RunsOn and Uses are missing
		return nil, fmt.Errorf("job MUST be contains either `runs-on` or `uses`")
	}
}

func (a *array) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format (e.g., "first")
	case jsontext.KindString:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			*a = []string{tok.String()}
			return nil
		}

	// 2. Standard array format (e.g., [first, second,...])
	case jsontext.KindBeginArray:
		type alias array
		return json.UnmarshalDecode(d, (*alias)(a))

	default:
		return fmt.Errorf("expected string or array, got kind %v", k)
	}
}

func (s *JobSecrets) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. "inherit" secret
	case jsontext.KindString:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else if t := tok.String(); t == "inherit" {
			s.Inherit = true
			return nil
		} else {
			return fmt.Errorf("unexpected JobSecrets=%v", t)
		}

	// 2. Standard object format (e.g., {"k": "v",...})
	case jsontext.KindBeginObject:
		return json.UnmarshalDecode(d, &s.Secrets)

	default:
		return fmt.Errorf("expected object for JobSecrets, got kind %v", k)
	}
}

func (e *Environment) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format (e.g., "first")
	case jsontext.KindString:
		return json.UnmarshalDecode(d, &e.Name)

	// 2. Standard object format (e.g., {"name": "prod", "url": "http://foobar.app"})
	case jsontext.KindBeginObject:
		type alias Environment
		return json.UnmarshalDecode(d, (*alias)(e))

	default:
		return fmt.Errorf("expected string or object for Environment, got kind %v", k)
	}
}

func (r *RunsOn) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string/[]string format (e.g., "first")
	case jsontext.KindString,
		jsontext.KindBeginArray:
		return json.UnmarshalDecode(d, &r.Labels)

	// 2. Full object format (e.g., {"group": "gpu-runner", "labels": ["label1", "label2",...]})
	case jsontext.KindBeginObject:
		type alias RunsOn
		return json.UnmarshalDecode(d, (*alias)(r))

	default:
		return fmt.Errorf("expected string, array or object for RunsOn, got kind %v", k)
	}
}
