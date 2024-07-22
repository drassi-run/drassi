package traits

import "drassi.run/core/pkg/expression/types/ref"

// Indexer permits random access of elements by index 'a[b()]'.
type Indexer interface {
	// IndexType returns data type of Index
	IndexType() ref.Type

	// Get the value at the specified index or error.
	Get(index any) (ref.Val, error)
}
