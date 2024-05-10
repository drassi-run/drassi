package gha

import (
	"encoding/json"
	"errors"
	"time"
)

var zeroTime time.Time

func init() {
	zeroTime, _ = time.Parse(time.TimeOnly, "00:00:00")
}

func (d *Duration) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		return nil, d.parse(s)
	} else {
		return input, nil
	}
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		return d.parse(value)
	default:
		return errors.New("invalid duration")
	}
}

func (d *Duration) parse(s string) error {
	// first try (go time.Duration format)
	if duration, err := time.ParseDuration(s); err == nil {
		d.Duration = duration
		return nil
	}

	// second try (C# Timespan format)
	if t, err := time.Parse(time.TimeOnly, s); err == nil {
		d.Duration = t.Sub(zeroTime)
		return nil
	}

	return errors.New("unknown duration format: " + s)
}

func (t *Time) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		return nil, t.parse(s)
	} else {
		return input, nil
	}
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		return t.parse(value)
	default:
		return errors.New("invalid time")
	}
}

func (t *Time) parse(s string) error {
	// first try (RFC3339 format with zone info)
	if dt, err := time.Parse(time.RFC3339, s); err == nil {
		t.Time = dt
		return nil
	}

	// second try (RFC3339 format w/o zone info)
	if dt, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		t.Time = dt
		return nil
	}

	return errors.New("unknown time format: " + s)
}
