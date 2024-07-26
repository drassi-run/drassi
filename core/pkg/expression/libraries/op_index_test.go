package libraries

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"testing"
)

type object struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var listString = []string{"zero", "one", "two", "three"}
var listInt = []int64{10, 11, 12, 13, 14, 15}
var listUint = []int64{10, 11, 12, 13, 14, 15} // dynamic list (reflection)
var listFloat = []float64{0.1, 1.7, 3.14}
var mapSS = map[string]string{"zero": "zeroth", "one": "first", "two": "second", "three": "third"}
var mapIS = map[int64]string{0: "zeroth", 1: "first", 2: "second", 3: "third"} // dynamic map (reflection)
var objectX = &object{Name: "drassi", Age: 0}
var objectY = &object{Name: "action", Age: 10}

func TestIndex(t *testing.T) {
	t.Run("normal", testIndexNormal)
	t.Run("not-existed", testIndexNotExisted)
	t.Run("wildcard", testIndexWildcard)
}

func testIndexNormal(t *testing.T) {
	testcases := []struct {
		object   any
		inputs   any
		expected any
	}{
		{listString, 0, "zero"},
		{listInt, 3, 13},
		{listFloat, 2, 3.14},
		{mapSS, "one", "first"},
		{mapIS, 3, "third"},

		{listUint, false, 10},  // false convert to 0
		{listInt, "5", 15},     // str to int
		{mapIS, true, "first"}, // true convert to 1
		{mapIS, "2", "second"}, // str to int
	}
	for _, tc := range testcases {
		target := toLazy(tc.object)
		index := toLazy(tc.inputs)

		actual := Index(target, index)
		verify(t, tc.expected, actual, "index(%v, %v)", tc.object, tc.inputs)
	}
}

func testIndexNotExisted(t *testing.T) {
	type empty struct{}

	testcases := []struct {
		object any
		inputs any
	}{
		// string key
		{listString, "not-existed"},
		{listInt, "not-existed"},
		{listUint, "not-existed"},
		{listFloat, "not-existed"},
		{mapSS, "not-existed"},
		{mapIS, "not-existed"},

		// int key
		{listString, -1},
		{listString, len(listString)},
		{listInt, math.MaxInt32},
		{listUint, math.MaxInt32},
		{listFloat, math.MaxInt32},
		{mapSS, math.MaxInt32},
		{mapIS, math.MaxInt32},

		// other kind
		{listString, empty{}},
		{listInt, empty{}},
		{listUint, empty{}},
		{listFloat, empty{}},
		{mapSS, empty{}},
		{mapIS, empty{}},
	}
	for _, tc := range testcases {
		target := toLazy(tc.object)
		index := toLazy(tc.inputs)

		actual := Index(target, index)
		verify(t, nil, actual, "index(%v, %v)", tc.object, tc.inputs)
	}
}

func testIndexWildcard(t *testing.T) {
	fruits := []map[string]any{
		{"name": "apple", "quantity": 1},
		{"name": "orange", "quantity": 2},
		{"name": "pear", "quantity": 1},
	}
	vegetables := map[string]map[string][]string{
		"scallions": {
			"colors":         []string{"green", "white", "red"},
			"ediblePortions": []string{"roots", "stalks"},
		},
		"beets": {
			"colors":         []string{"purple", "red", "gold", "white", "pink"},
			"ediblePortions": []string{"roots", "stems", "leaves"},
		},
		"artichokes": {
			"colors":         []string{"green", "purple", "red", "black"},
			"ediblePortions": []string{"hearts", "stems", "leaves"},
		},
	}

	t.Run("list-of-map", func(t *testing.T) {
		indexes := []any{"*", "name"} // fruits.*.name
		expected := []any{"apple", "orange", "pear"}
		runWildcardTest(t, fruits, indexes, expected)
	})

	t.Run("map-of-map", func(t *testing.T) {
		indexes := []any{"*", "ediblePortions"} // vegetables.*.ediblePortions
		expected := []any{
			[]string{"roots", "stalks"},
			[]string{"roots", "stems", "leaves"},
			[]string{"hearts", "stems", "leaves"},
		}
		runWildcardTest(t, vegetables, indexes, expected)

		indexes = append(indexes, 0) // vegetables.*.ediblePortions.0
		expected = []any{"roots", "roots", "hearts"}
		runWildcardTest(t, vegetables, indexes, expected)
	})

	t.Run("nested/1", func(t *testing.T) {
		indexes := []any{"*", "*"} // fruits.*.*
		expected := []any{"apple", int64(1), "orange", int64(1), "pear", int64(2)}
		runWildcardTest(t, fruits, indexes, expected)
	})

	t.Run("nested/2", func(t *testing.T) {
		indexes := []any{"*", "*"} // vegetables.*.*
		expected := []any{
			[]string{"green", "white", "red"},
			[]string{"roots", "stalks"},
			[]string{"purple", "red", "gold", "white", "pink"},
			[]string{"roots", "stems", "leaves"},
			[]string{"green", "purple", "red", "black"},
			[]string{"hearts", "stems", "leaves"},
		}
		runWildcardTest(t, vegetables, indexes, expected)
	})

	t.Run("nested/3", func(t *testing.T) {
		indexes := []any{"*", "*", "0"} // vegetables.*.*[0]
		expected := []any{"green", "roots", "purple", "roots", "green", "hearts"}
		runWildcardTest(t, vegetables, indexes, expected)
	})

	t.Run("nested/4", func(t *testing.T) {
		indexes := []any{"*", "*", "3"} // vegetables.*.*[3]
		expected := []any{"white", "black"}
		runWildcardTest(t, vegetables, indexes, expected)
	})

	t.Run("nested/5", func(t *testing.T) {
		indexes := []string{"*", "*", "*"} // fruits.*.*.*
		expected := []string{}             // empty
		runWildcardTest(t, fruits, indexes, expected)

		indexes = []string{"*", "*", "*", "*", "*", "*"}
		runWildcardTest(t, fruits, indexes, expected)
	})
}

func runWildcardTest[O any, I any, E any](t *testing.T, object O, indexes []I, expected E) {
	actual := Index(toLazy(object), toLazies(indexes...)...)

	err, _ := actual.(error)
	assert.NoError(t, err, indexes)
	assert.True(t, compareSet(expected, actual.Value()), indexes)
}

func compareSet(x, y any) bool {
	vx := reflect.ValueOf(x)
	if kind := vx.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return assert.ObjectsAreEqualValues(x, y)
	}

	vy := reflect.ValueOf(y)
	if kind := vy.Kind(); kind != reflect.Array && kind != reflect.Slice {
		return assert.ObjectsAreEqualValues(x, y)
	}

	if vx.Len() != vy.Len() {
		return false
	}

	sx := listToDict(vx)
	sy := listToDict(vy)
	return assert.ObjectsAreEqualValues(sx, sy)
}

func listToDict(v reflect.Value) map[string]any {
	m := make(map[string]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)
		if !e.IsValid() {
			continue
		}

		item := e.Interface()
		if item == nil {
			continue
		}

		if b, err := json.Marshal(item); err == nil {
			key := string(b)
			m[key] = item
		}
	}

	return m
}
