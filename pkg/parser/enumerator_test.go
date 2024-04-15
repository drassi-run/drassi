package parser

import (
	"testing"
)

func TestNewEnumerator(t *testing.T) {
	values := []interface{}{1, "two", 3.0}
	enum := NewEnumerator(values...)

	if len(enum.values) != len(values) {
		t.Errorf("NewEnumerator() should have %d values, got %d", len(values), len(enum.values))
	}

	if enum.index != -1 {
		t.Errorf("NewEnumerator() index should be -1, got %d", enum.index)
	}
}

func TestEnumerator_Next(t *testing.T) {
	values := []interface{}{1, "two", 3.0}
	enum := NewEnumerator(values...)

	// Test Next() method
	for _, expectedValue := range values {
		if !enum.Next() {
			t.Errorf("Next() should return true until the end of the slice")
		}
		actualValue := enum.Value()
		if actualValue != expectedValue {
			t.Errorf("Expected value %v, got %v", expectedValue, actualValue)
		}
	}

	// Test that Next() returns false at the end of the slice
	if enum.Next() {
		t.Errorf("Next() should return false when there are no more elements")
	}
}

func TestEnumerator_Value(t *testing.T) {
	tests := []struct {
		values    []interface{}
		wantValue interface{}
	}{
		{[]interface{}{1, "two", 3.0}, 1},
		{[]interface{}{"only one"}, "only one"},
		{[]interface{}{}, nil},
	}

	for _, tt := range tests {
		enum := NewEnumerator(tt.values...)
		gotValue := enum.Value()
		if gotValue != nil {
			t.Errorf("Value() before calling Next() should be nil, got %v", gotValue)
		}
		if enum.Next() {
			gotValue = enum.Value()
			if gotValue != tt.wantValue {
				t.Errorf("Expected enumerator to return value %v, got %v", tt.wantValue, gotValue)
			}
		} else if len(tt.values) > 0 {
			t.Errorf("Next() should return true but got false")
		}
	}
}
