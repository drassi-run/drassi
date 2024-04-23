package shared

// Enumerator represents a type that can iterate over a collection of values.
type Enumerator struct {
	values []interface{} // Slice to hold the values
	index  int           // Current index
}

// NewEnumerator creates a new instance of Enumerator with the given values.
func NewEnumerator(values ...interface{}) *Enumerator {
	return &Enumerator{values: values, index: -1} // Initialize index to -1
}

// Next moves the enumerator to the next element in the collection.
// It returns false when there are no more elements to iterate.
func (e *Enumerator) Next() bool {
	e.index++
	return e.index < len(e.values)
}

// Value returns the current value of the enumerator.
func (e *Enumerator) Value() interface{} {
	if e.index < 0 || e.index >= len(e.values) {
		return nil
	}
	return e.values[e.index]
}
