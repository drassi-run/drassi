package traits

import "drassi.run/core/pkg/expression/types/ref"

// Iterable aggregate types permit traversal over their elements.
type Iterable interface {
	// Iterator returns a new iterator view of the struct.
	Iterator() Iterator
}

// Iterator permits safe traversal over the contents of an aggregate type.
type Iterator interface {
	// HasNext returns true if there are unvisited elements in the Iterator.
	HasNext() bool

	// Next returns the next element.
	Next() ref.Val
}
