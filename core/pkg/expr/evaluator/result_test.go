package evaluator

import (
	"math"
	"testing"
)

func TestAbstractEqual(t *testing.T) {
	t.Run("Compare Null Values", func(t *testing.T) {
		if !equal(nil, nil) {
			t.Error("Expected two null values to be equal")
		}
	})

	t.Run("Compare Numbers", func(t *testing.T) {
		if !equal(42.0, 42.0) {
			t.Error("Expected two identical numbers to be equal")
		}
		if equal(42.0, 43.0) {
			t.Error("Expected two different numbers not to be equal")
		}
		if equal(math.NaN(), math.NaN()) {
			t.Error("Expected two NaNs not to be equal")
		}
	})

	t.Run("Compare Strings", func(t *testing.T) {
		if !equal("Go", "go") {
			t.Error("Expected two strings with different case to be equal")
		}
		if equal("Go", "Golang") {
			t.Error("Expected two different strings not to be equal")
		}
	})

	t.Run("Compare Booleans", func(t *testing.T) {
		if !equal(true, true) {
			t.Error("Expected two 'true' values to be equal")
		}
		if equal(true, false) {
			t.Error("Expected 'true' and 'false' not to be equal")
		}
	})

	t.Run("Compare Objects", func(t *testing.T) {
		obj1 := map[string]any{"key": "value"}
		obj2 := map[string]any{"key": "value"}
		obj3 := obj1
		obj4 := obj1
		if !equal(obj3, obj4) {
			t.Error("Expected two references to the same object to be equal")
		}
		if equal(obj1, obj2) {
			t.Error("Expected two separate but identical objects not to be considered equal")
		}
		if !equal(obj1, obj3) {
			t.Error("Expected object and references to the it to be equal")
		}
	})

	t.Run("Compare Slices/Arrays", func(t *testing.T) {
		arr1 := []any{1, 2, 3}
		arr2 := []any{1, 2, 3}
		arr3 := arr1
		arr4 := [3]int{1, 2, 3}
		arr5 := arr4
		arr6 := arr4
		if !equal(arr1, arr3) {
			t.Error("Expected two slices reference to the same underlying array to be equal")
		}
		if equal(arr1, arr2) {
			t.Error("Expected two separate but identical slices not to be considered equal")
		}
		if !equal(arr4, arr5) {
			t.Error("Expected two separate but identical arrays to be considered equal")
		}
		if !equal(arr5, arr6) {
			t.Error("Expected two separate but identical arrays to be considered equal")
		}
	})
	t.Run("Mismatched Types", func(t *testing.T) {
		if equal(42, 42) {
			t.Error("Expect non-zero int not equal")
		}
		if !equal(0, 0) {
			t.Error("Expect zero value equal")
		}
		if !equal(struct{}{}, struct{}{}) {
			t.Error("Expect not int equal")
		}
		if equal(42, "42") {
			t.Error("Expected integer number and string not to be equal")
		}
		if !equal(42.0, "42") {
			t.Error("Expected float number and string to be equal")
		}
		if !equal(42.0, 42.0) {
			t.Error("Expected float number to be equal")
		}
		if equal(true, "true") {
			t.Error("Expected boolean and string not to be equal")
		}
	})
}
